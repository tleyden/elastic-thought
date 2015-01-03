package elasticthought

import (
	"fmt"
	"sync"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// A classify job tries to classify images given by user against
// the given trained model
type ClassifyJob struct {
	ElasticThoughtDoc
	ProcessingState ProcessingState `json:"processing-state"`
	ClassifierID    string          `json:"classifier-id"`

	// had to make exported, due to https://github.com/gin-gonic/gin/pull/123
	// waiting for this to get merged into master branch, since go get
	// pulls from master branch.
	Configuration Configuration
}

// Create a new classify job.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewClassifyJob(c Configuration) *ClassifyJob {
	return &ClassifyJob{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_CLASSIFY_JOB},
		Configuration:     c,
	}
}

// Run this job
func (c *ClassifyJob) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	logg.LogTo("CLASSIFY_JOB", "Run() called!")

	updatedState, err := c.UpdateProcessingState(Processing)
	if err != nil {

		// TODO: c.recordProcessingError(err)
		return
	}

	if !updatedState {
		logg.LogTo("CLASSIFY_JOB", "%+v already processed.  Ignoring.", c)
		return
	}

	// TODO: refactor to use new generic cas approach

	// TODO: add code to run job

	// lazily create dir and download prototxt if doesn't exist

	// invoke caffe

	// extract results

	// update classifyjob with results

}

// Update the processing state to new state.
func (c *ClassifyJob) UpdateProcessingState(newState ProcessingState) (bool, error) {

	updater := func(classifyJob *ClassifyJob) {
		classifyJob.ProcessingState = newState
	}

	doneMetric := func(classifyJob ClassifyJob) bool {
		return classifyJob.ProcessingState == newState
	}

	return c.casUpdate(updater, doneMetric)

}

// The first return value will be true when it was updated due to calling this method,
// or false if it was already in that state or put in that state by something else
// during the update attempt.
//
// If any errors occur while trying to update, they will be returned in the second
// return value.
func (c *ClassifyJob) casUpdate(updater func(*ClassifyJob), doneMetric func(ClassifyJob) bool) (bool, error) {

	db := c.Configuration.DbConnection()

	genUpdater := func(classifyJobPtr interface{}) {
		cjp := classifyJobPtr.(*ClassifyJob)
		updater(cjp)
	}

	genDoneMetric := func(classifyJobPtr interface{}) bool {
		cjp := classifyJobPtr.(*ClassifyJob)
		return doneMetric(*cjp)
	}

	refresh := func(classifyJobPtr interface{}) error {
		cjp := classifyJobPtr.(*ClassifyJob)
		return cjp.RefreshFromDB(db)
	}

	return casUpdate(db, c, genUpdater, genDoneMetric, refresh)

}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (c *ClassifyJob) Insert() error {

	db := c.Configuration.DbConnection()

	id, rev, err := db.Insert(c)
	if err != nil {
		err := fmt.Errorf("Error inserting classify job: %v.  Err: %v", c, err)
		return err
	}

	c.Id = id
	c.Revision = rev

	return nil

}

// CodeReview: duplication with RefreshFromDB in many places
func (c *ClassifyJob) RefreshFromDB(db couch.Database) error {
	classifyJob := ClassifyJob{}
	err := db.Retrieve(c.Id, &classifyJob)
	if err != nil {
		return err
	}
	*c = classifyJob
	return nil
}

// Find a classify Job in the db with the given id, or error if not found
// CodeReview: duplication with Find in many places
func (c *ClassifyJob) Find(id string) error {
	db := c.Configuration.DbConnection()
	c.Id = id
	if err := c.RefreshFromDB(db); err != nil {
		return err
	}
	return nil
}
