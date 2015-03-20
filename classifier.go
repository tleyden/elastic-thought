package elasticthought

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/couchbaselabs/logg"
	"github.com/golang/protobuf/proto"
	"github.com/tleyden/elastic-thought/caffe"
	"github.com/tleyden/go-couch"
)

// A classifier uses a trained model to classify new incoming data points
type Classifier struct {
	ElasticThoughtDoc
	SpecificationUrl string `json:"specification-url" binding:"required"`
	TrainingJobID    string `json:"training-job-id" binding:"required"`
	Scale            string `json:"scale" binding:"required"`
	ImageWidth       string `json:"image-width" binding:"required"`
	ImageHeight      string `json:"image-height" binding:"required"`
	Color            bool   `json:"color"`
	Gpu              bool   `json:"gpu"`

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

	if _, err := c.casUpdate(updater, doneMetric); err != nil {
		return err
	}

	return nil

}

func (c *Classifier) casUpdate(updater func(*Classifier), doneMetric func(Classifier) bool) (bool, error) {

	db := c.Configuration.DbConnection()

	genUpdater := func(classifierPtr interface{}) {
		cjp := classifierPtr.(*Classifier)
		updater(cjp)
	}

	genDoneMetric := func(classifierPtr interface{}) bool {
		cjp := classifierPtr.(*Classifier)
		return doneMetric(*cjp)
	}

	refresh := func(classifierPtr interface{}) error {
		cjp := classifierPtr.(*Classifier)
		return cjp.RefreshFromDB(db)
	}

	return casUpdate(db, c, genUpdater, genDoneMetric, refresh)

}

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

// Find a classifier in the db with the given id,
// or return an error if not found
func (c *Classifier) Find(id string) error {
	db := c.Configuration.DbConnection()
	c.Id = id
	if err := c.RefreshFromDB(db); err != nil {
		return err
	}
	return nil
}

func (c Classifier) Validate() error {

	logg.LogTo("CLASSIFIER", "Validate training job")
	if err := c.validateTrainingJob(); err != nil {
		return err
	}

	logg.LogTo("CLASSIFIER", "Validate classifier net")
	if err := c.validateClassifierNet(); err != nil {
		return err
	}

	logg.LogTo("CLASSIFIER", "Validation passed")

	return nil
}

func (c Classifier) validateTrainingJob() error {

	trainingJob := NewTrainingJob(c.Configuration)

	err := trainingJob.Find(c.TrainingJobID)
	if err != nil {
		return err
	}

	return nil

}

// make sure the specification url points to a valid prototxt file
func (c Classifier) validateClassifierNet() error {

	_, err := c.classifierNet()
	if err != nil {
		return err
	}
	return nil

}

// read the classifier prototxt and create protobuf struct and return
func (c Classifier) classifierNet() (*caffe.NetParameter, error) {

	specContents, err := c.getClassifierPrototxt()
	if err != nil {
		return nil, fmt.Errorf("Error getting classifier prototxt content.  Err: %v", err)
	}

	// read into object with protobuf (must have already generated go protobuf code)
	netParam := &caffe.NetParameter{}

	if err := proto.UnmarshalText(string(specContents), netParam); err != nil {
		logg.LogTo("CLASSIFIER", "Error unmarshalling %v", string(specContents))
		return nil, err
	}

	return netParam, nil

}

// read raw classifier prototxt from url and return bytes
func (c Classifier) getClassifierPrototxt() ([]byte, error) {

	resp, err := http.Get(c.SpecificationUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil

}

func (c Classifier) getTrainingJob() (*TrainingJob, error) {

	trainingJob := NewTrainingJob(c.Configuration)
	err := trainingJob.Find(c.TrainingJobID)
	return trainingJob, err

}
