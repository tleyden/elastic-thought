package elasticthought

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// A changes listener listens for changes on the _changes feed and reacts to them.
// The changes listener currently runs as a goroutine in the httpd process, and
// so the system only currently supports having a single httpd process, because otherwise
// there would be multiple changes listeners on the same changes feed, which will
// cause duplicate jobs to get kicked off.  If the system needs to support multiple
// http processes, then the changes listener needs to run in its own process.
type ChangesListener struct {
	Configuration Configuration
	Database      couch.Database
	JobScheduler  JobScheduler
}

// Create a new ChangesListener
func NewChangesListener(c Configuration, jobScheduler JobScheduler) (*ChangesListener, error) {

	db := c.DbConnection()

	return &ChangesListener{
		Configuration: c,
		Database:      db,
		JobScheduler:  jobScheduler,
	}, nil
}

// Follow changes feed.  This will typically be run in its own goroutine.
func (c ChangesListener) FollowChangesFeed() {

	logg.LogTo("CHANGES", "going to follow changes feed")

	var since interface{}

	handleChange := func(reader io.Reader) interface{} {
		logg.LogTo("CHANGES", "handleChange() callback called")
		changes, err := decodeChanges(reader)
		if err != nil {
			// it's very common for this to timeout while waiting for new changes.
			// since we want to follow the changes feed forever, just log an error
			logg.LogTo("CHANGES", "%T decoding changes: %v.", err, err)
			return since
		}
		c.processChanges(changes)

		since = changes.LastSequence
		logg.LogTo("CHANGES", "returning since: %v", since)
		return since

	}

	options := map[string]interface{}{}
	options["feed"] = "longpoll"

	logg.LogTo("CHANGES", "Following changes feed: %+v.", options)

	// this will block until the handleChange callback returns nil
	c.Database.Changes(handleChange, options)

	logg.LogPanic("Changes listener died -- this should never happen")

}

func (c ChangesListener) processChanges(changes couch.Changes) {

	for _, change := range changes.Results {

		if change.Deleted {
			logg.LogTo("CHANGES", "change was deleted, skipping")
			continue
		}

		// ignore certain docs, like "_user/*"
		if strings.HasPrefix(change.Id, "_user") {
			logg.LogTo("CHANGES", "Ignoring change: %v", change.Id)
			continue
		}

		doc := ElasticThoughtDoc{}
		err := c.Database.Retrieve(change.Id, &doc)

		if err != nil {
			errMsg := fmt.Errorf("Didn't retrieve: %v - %v", change.Id, err)
			logg.LogError(errMsg)
			continue
		}

		switch doc.Type {
		case DOC_TYPE_DATAFILE:
			c.handleDatafileChange(change, doc)
		case DOC_TYPE_DATASET:
			c.handleDatasetChange(change, doc)
		case DOC_TYPE_TRAINING_JOB:
			c.handleTrainingJobChange(change, doc)
		case DOC_TYPE_CLASSIFY_JOB:
			c.handleClassifyJobChange(change, doc)

		}

	}

}

func (c ChangesListener) handleTrainingJobChange(change couch.Change, doc ElasticThoughtDoc) {

	logg.LogTo("CHANGES", "got a training job doc: %+v", doc)

	// create a Training Job doc from the ElasticThoughtDoc
	trainingJob := &TrainingJob{}
	if err := c.Database.Retrieve(change.Id, &trainingJob); err != nil {
		errMsg := fmt.Errorf("Didn't retrieve: %v - %v", change.Id, err)
		logg.LogError(errMsg)
		return
	}

	// check the state, only schedule if state == pending
	if trainingJob.ProcessingState != Pending {
		logg.LogTo("CHANGES", "State != pending: %+v", trainingJob)
		return
	}

	job := NewJobDescriptor(doc.Id)
	c.JobScheduler.ScheduleJob(*job)

}

func (c ChangesListener) handleClassifyJobChange(change couch.Change, doc ElasticThoughtDoc) {

	logg.LogTo("CHANGES", "got a classify job doc: %+v", doc)

	// create a Training Job doc from the ElasticThoughtDoc
	classifyJob := NewClassifyJob(c.Configuration)
	if err := classifyJob.Find(change.Id); err != nil {
		errMsg := fmt.Errorf("Could not find: %v - %v", change.Id, err)
		logg.LogError(errMsg)
		return
	}

	// check the state, only schedule if state == pending
	if classifyJob.ProcessingState != Pending {
		logg.LogTo("CHANGES", "State != pending: %+v", classifyJob)
		return
	}

	job := NewJobDescriptor(doc.Id)
	c.JobScheduler.ScheduleJob(*job)

}

func (c ChangesListener) handleDatasetChange(change couch.Change, doc ElasticThoughtDoc) {

	logg.LogTo("CHANGES", "got a dataset doc: %+v", doc)

	// create a Dataset doc from the ElasticThoughtDoc
	dataset := NewDataset(c.Configuration)
	if err := c.Database.Retrieve(change.Id, &dataset); err != nil {
		errMsg := fmt.Errorf("Didn't retrieve: %v - %v", change.Id, err)
		logg.LogError(errMsg)
		return
	}

	logg.LogTo("CHANGES", "convert to dataset: %+v", dataset)

	// check the state, only schedule if state == pending
	if dataset.ProcessingState != Pending {
		logg.LogTo("CHANGES", "Dataset state != pending: %+v", dataset)
		return
	}

	job := NewJobDescriptor(doc.Id)
	c.JobScheduler.ScheduleJob(*job)

}

func (c ChangesListener) handleDatafileChange(change couch.Change, doc ElasticThoughtDoc) {

	logg.LogTo("CHANGES", "got a datafile doc: %+v", doc)

	// create a Datafile doc from the ElasticThoughtDoc
	datafile := NewDatafile(c.Configuration)
	if err := c.Database.Retrieve(change.Id, &datafile); err != nil {
		errMsg := fmt.Errorf("Didn't retrieve: %v - %v", change.Id, err)
		logg.LogError(errMsg)
		return
	}

	logg.LogTo("CHANGES", "convert to datafile: %+v", datafile)

	// check the state, only schedule if state == pending
	if datafile.ProcessingState != Pending {
		logg.LogTo("CHANGES", "Datafile state != pending: %+v", datafile)
		return
	}

	job := NewJobDescriptor(doc.Id)
	c.JobScheduler.ScheduleJob(*job)

}

func decodeChanges(reader io.Reader) (couch.Changes, error) {

	changes := couch.Changes{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&changes)
	if err != nil {
		logg.LogTo("CHANGES", "Err decoding changes: %v", err)
	}
	return changes, err

}
