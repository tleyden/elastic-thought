package elasticthought

import (
	"fmt"

	"github.com/tleyden/go-couch"
)

/*
A dataset is created from a datafile, and represents a partition of the datafile
to be used for a particular purpose.  The typical example would involve:
    - Datafile with 100 examples
    - Training dataset with 70 examples
    - Test dataset with 30 examples
*/
type Dataset struct {
	ElasticThoughtDoc
	DatafileID      string          `json:"datafile-id" binding:"required"`
	ProcessingState ProcessingState `json:"processing-state"`
	TrainingDataset TrainingDataset `json:"training" binding:"required"`
	TestDataset     TestDataset     `json:"test" binding:"required"`
	ProcessingLog   string          `json:"processing-log"`
}

type TrainingDataset struct {
	SplitPercentage float64 `json:"split-percentage"`
	Url             string  `json:"url"`
}

type TestDataset struct {
	SplitPercentage float64 `json:"split-percentage"`
	Url             string  `json:"url"`
}

// Create a new dataset.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewDataset() *Dataset {
	return &Dataset{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_DATASET},
	}
}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (d Dataset) Insert(db couch.Database) (*Dataset, error) {

	id, _, err := db.Insert(d)
	if err != nil {
		err := fmt.Errorf("Error inserting dataset: %v.  Err: %v", d, err)
		return nil, err
	}

	// load dataset object from db (so we have id/rev fields)
	dataset := &Dataset{}
	err = db.Retrieve(id, dataset)
	if err != nil {
		err := fmt.Errorf("Error fetching dataset: %v.  Err: %v", id, err)
		return nil, err
	}

	return dataset, nil

}

// Find and return the datafile associated with this dataset
func (d Dataset) GetDatafile(db couch.Database) (*Datafile, error) {
	datafile := &Datafile{}
	if err := db.Retrieve(d.DatafileID, datafile); err != nil {
		return nil, err
	}
	return datafile, nil
}

// Update the dataset state to record that it finished successfully
func (d Dataset) FinishedSuccessfully(db couch.Database) error {

	d.ProcessingState = FinishedSuccessfully

	// TODO: retry if 409 error
	_, err := db.Edit(d)

	if err != nil {
		return err
	}

	return nil

}

// Update the dataset state to record that it failed
func (d Dataset) Failed(db couch.Database, processingErr error) error {

	d.ProcessingState = Failed
	d.ProcessingLog = fmt.Sprintf("%v", processingErr)

	// TODO: retry if 409 error
	_, err := db.Edit(d)

	if err != nil {
		return err
	}

	return nil

}

// Path to training artifact file, eg <id>/training.tar.gz
func (d Dataset) TrainingArtifactPath() string {
	return fmt.Sprintf("%v/%v", d.Id, TRAINING_ARTIFACT)
}

// Path to testing artifact file, eg <id>/testing.tar.gz
func (d Dataset) TestingArtifactPath() string {
	return fmt.Sprintf("%v/%v", d.Id, TEST_ARTIFACT)
}

// Update this dataset with the artifact urls (cbfs://<id>/training.tar.gz, ..)
// even though these artifacts might not exist yet.
func (d Dataset) AddArtifactUrls(db couch.Database) (*Dataset, error) {

	d.TrainingDataset.Url = fmt.Sprintf("%v%v", CBFS_URI_PREFIX, d.TrainingArtifactPath())
	d.TestDataset.Url = fmt.Sprintf("%v%v", CBFS_URI_PREFIX, d.TestingArtifactPath())

	// TODO: retry if 409 error
	_, err := db.Edit(d)

	if err != nil {
		return nil, err
	}

	// load latest version of dataset to return
	dataset := &Dataset{}
	err = db.Retrieve(d.Id, dataset)
	if err != nil {
		return nil, err
	}

	return dataset, nil

}
