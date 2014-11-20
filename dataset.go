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

// Create a new dataset
func NewDataset() *Dataset {
	return &Dataset{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_DATASET},
	}
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
	return fmt.Sprintf("%v/training.tar.gz", d.Id)
}

// Path to testing artifact file, eg <id>/testing.tar.gz
func (d Dataset) TestingArtifactPath() string {
	return fmt.Sprintf("%v/testing.tar.gz", d.Id)
}

// Update this dataset with the artifact urls (cbfs://<id>/training.tar.gz, ..)
// even though these artifacts might not exist yet.
func (d Dataset) AddArtifactUrls(db couch.Database) (Dataset, error) {

	d.TrainingDataset.Url = fmt.Sprintf("cbfs://%v", d.TrainingArtifactPath())
	d.TestDataset.Url = fmt.Sprintf("cbfs://%v", d.TestingArtifactPath())

	// TODO: retry if 409 error
	_, err := db.Edit(d)

	if err != nil {
		return Dataset{}, err
	}

	// load latest version of dataset to return
	dataset := Dataset{}
	err = db.Retrieve(d.Id, &dataset)
	if err != nil {
		return Dataset{}, err
	}

	return dataset, nil

}
