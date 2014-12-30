package elasticthought

import "fmt"

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
func NewClassifier() *Classifier {
	return &Classifier{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_CLASSIFIER},
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
