package elasticthought

import (
	"fmt"
	"sync"

	"github.com/couchbaselabs/logg"
	"github.com/dustin/httputil"
)

// A classify job tries to classify images given by user against
// the given trained model
type ClassifyJob struct {
	ElasticThoughtDoc
	ProcessingState ProcessingState `json:"processing-state"`

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
//
// CodeReview: major duplication with trainingJob.casUpdate
func (c *ClassifyJob) casUpdate(updater func(*ClassifyJob), doneMetric func(ClassifyJob) bool) (bool, error) {

	db := c.Configuration.DbConnection()

	// if already has the newState, return false
	if doneMetric(*c) == true {
		logg.LogTo("CLASSIFY_JOB", "Already has new state, nothing to do: %+v", c)
		return false, nil
	}

	for {
		updater(c)

		// SAVE: try to save to the database
		logg.LogTo("CLASSIFY_JOB", "Trying to save: %+v", c)

		_, err := db.Edit(c)

		if err != nil {

			logg.LogTo("CLASSIFY_JOB", "Got error updating: %v", err)

			// if it failed with any other error than 409, return an error
			if !httputil.IsHTTPStatus(err, 409) {
				logg.LogTo("CLASSIFY_JOB", "Not a 409 error: %v", err)
				return false, err
			}

			// it failed with 409 error
			logg.LogTo("CLASSIFY_JOB", "Its a 409 error: %v", err)

			// get the latest version of the document
			if err := c.RefreshFromDB(); err != nil {
				return false, err
			}

			logg.LogTo("CLASSIFY_JOB", "Retrieved new: %+v", c)

			// does it already have the new the state (eg, someone else set it)?
			if doneMetric(*c) == true {
				logg.LogTo("CLASSIFY_JOB", "doneMetric returned true, nothing to do")
				return false, nil
			}

			// no, so try updating state and saving again
			continue

		}

		// successfully saved, we are done
		logg.LogTo("CLASSIFY_JOB", "Successfully saved: %+v", c)
		return true, nil

	}

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
func (c *ClassifyJob) RefreshFromDB() error {
	db := c.Configuration.DbConnection()
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
	c.Id = id
	if err := c.RefreshFromDB(); err != nil {
		return err
	}
	return nil
}
