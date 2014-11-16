package elasticthought

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// A changes listener listens for changes on the _changes feed and reacts to them
type ChangesListener struct {
	Configuration Configuration
	Database      couch.Database
	JobRunner     JobRunner
}

// Create a new ChangesListener
func NewChangesListener(c Configuration, jobRunner JobRunner) (*ChangesListener, error) {

	db, err := couch.Connect(c.DbUrl)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error %v | dbUrl: %v", err, c.DbUrl))
		logg.LogError(err)
		return nil, err
	}

	return &ChangesListener{
		Configuration: c,
		Database:      db,
		JobRunner:     jobRunner,
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

	logg.LogTo("CHANGES", "Following changes feed: %+v", options)

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

		doc := ElasticThoughtDoc{}
		err := c.Database.Retrieve(change.Id, &doc)
		if err != nil {
			errMsg := fmt.Errorf("Didn't retrieve: %v - %v", change.Id, err)
			logg.LogError(errMsg)
			continue
		}

		switch doc.Type {
		case DOC_TYPE_DATASET:
			logg.LogTo("CHANGES", "got a dataset doc: %+v", doc)
			job := NewJobDescriptor(c.Configuration, doc.Id)
			c.JobRunner.ScheduleJob(*job)
		}

	}

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
