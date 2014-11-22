package elasticthought

import (
	"fmt"

	"github.com/tleyden/go-couch"
)

// A training job represents a "training session" of a solver against training/test data
type TrainingJob struct {
	ElasticThoughtDoc
	UserID          string          `json:"user-id"`
	ProcessingState ProcessingState `json:"processing-state"`
	SolverId        string          `json:"solver-id" binding:"required"`
}

// Create a new training job.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewTrainingJob() *TrainingJob {
	return &TrainingJob{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_TRAINING_JOB},
	}
}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (j TrainingJob) Insert(db couch.Database) (*TrainingJob, error) {

	id, _, err := db.Insert(j)
	if err != nil {
		err := fmt.Errorf("Error inserting training job: %+v.  Err: %v", j, err)
		return nil, err
	}

	// load dataset object from db (so we have id/rev fields)
	trainingJob := &TrainingJob{}
	err = db.Retrieve(id, trainingJob)
	if err != nil {
		err := fmt.Errorf("Error fetching training job: %v.  Err: %v", id, err)
		return nil, err
	}

	return trainingJob, nil

}
