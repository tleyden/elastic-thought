package elasticthought

import (
	"fmt"

	"github.com/couchbaselabs/logg"
	"github.com/dustin/httputil"
	"github.com/tleyden/go-couch"
)

// A classifier uses a trained model to classify new incoming data points
type Classifier struct {
	ElasticThoughtDoc
	SpecificationUrl string `json:"specification-url" binding:"required"`
	TrainingJobID    string `json:"training-job-id" binding:"required"`

	// had to make exported, due to https://github.com/gin-gonic/gin/pull/123
	// waiting for this to get merged into master branch, since go get
	// pulls from master branch.
	Configuration Configuration
}

// Create a new classifier.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewClassifier(c Configuration) *Classifier {
	return &Classifier{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_CLASSIFIER},
		Configuration:     c,
	}
}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (c *Classifier) Insert() error {

	db := c.Configuration.DbConnection()

	id, rev, err := db.Insert(c)
	if err != nil {
		err := fmt.Errorf("Error inserting classifier: %v.  Err: %v", c, err)
		return err
	}

	c.Id = id
	c.Revision = rev

	return nil

}

func (c *Classifier) SetSpecificationUrl(specUrlCbfs string) error {

	updater := func(classifier *Classifier) {
		classifier.SpecificationUrl = specUrlCbfs
	}

	doneMetric := func(classifier Classifier) bool {
		return classifier.SpecificationUrl == specUrlCbfs
	}

	if err := c.casUpdate(updater, doneMetric); err != nil {
		return err
	}

	return nil

}

// CodeReview: major duplication with trainingJob.casUpdate
func (c *Classifier) casUpdate(updater func(*Classifier), doneMetric func(Classifier) bool) error {

	db := c.Configuration.DbConnection()

	// if already has the newState, return false
	if doneMetric(*c) == true {
		logg.LogTo("CLASSIFIER", "Already has new state, nothing to do: %+v", c)
		return nil
	}

	for {
		updater(c)

		// SAVE: try to save to the database
		logg.LogTo("CLASSIFIER", "Trying to save: %+v", c)

		_, err := db.Edit(c)

		if err != nil {

			logg.LogTo("CLASSIFIER", "Got error updating: %v", err)

			// if it failed with any other error than 409, return an error
			if !httputil.IsHTTPStatus(err, 409) {
				logg.LogTo("CLASSIFIER", "Not a 409 error: %v", err)
				return err
			}

			// it failed with 409 error
			logg.LogTo("CLASSIFIER", "Its a 409 error: %v", err)

			// get the latest version of the document

			if err := c.RefreshFromDB(db); err != nil {
				return err
			}

			logg.LogTo("CLASSIFIER", "Retrieved new: %+v", c)

			// does it already have the new the state (eg, someone else set it)?
			if doneMetric(*c) == true {
				logg.LogTo("CLASSIFIER", "doneMetric returned true, nothing to do")
				return nil
			}

			// no, so try updating state and saving again
			continue

		}

		// successfully saved, we are done
		logg.LogTo("CLASSIFIER", "Successfully saved: %+v", c)
		return nil

	}

}

// CodeReview: duplication with trainingJob.casUpdate
func (c *Classifier) RefreshFromDB(db couch.Database) error {
	classifier := Classifier{}
	err := db.Retrieve(c.Id, &classifier)
	if err != nil {
		logg.LogTo("CLASSIFIER", "Error getting latest: %v", err)
		return err
	}
	*c = classifier
	return nil
}

func (c Classifier) Validate() error {

	trainingJob := NewTrainingJob(c.Configuration)

	err := trainingJob.Find(c.TrainingJobID)
	if err != nil {
		return err
	}
	return nil

}
