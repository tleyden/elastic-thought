package elasticthought

import "fmt"

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
