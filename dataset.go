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
	ProcessingState ProcessingState `json:"processing-state"`
	TrainingDataset TrainingDataset `json:"training" binding:"required"`
	TestDataset     TestDataset     `json:"test" binding:"required"`
	ProcessingLog   string          `json:"processing-log"`
}

type TrainingDataset struct {
	DatafileID      string  `json:"datafile-id" binding:"required"`
	SplitPercentage float64 `json:"split-percentage"`
	Url             string  `json:"url"`
}

type TestDataset struct {
	DatafileID      string  `json:"datafile-id" binding:"required"`
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
func (d Dataset) GetSplittableDatafile(db couch.Database) (*Datafile, error) {
	if !d.isSplittable() {
		return nil, fmt.Errorf("This dataset is not splittable")
	}

	// if its splittable, then the trainingset and the testset should have the
	// same datafile id
	if d.TrainingDataset.DatafileID != d.TestDataset.DatafileID {
		return nil, fmt.Errorf("Datafile id's for this dataset don't match")
	}

	// choose either datafile id since they are the same
	datafileId := d.TrainingDataset.DatafileID

	datafile := &Datafile{}
	if err := db.Retrieve(datafileId, datafile); err != nil {
		return nil, err
	}
	return datafile, nil
}

// Is this dataset splittable or has it already been split?
func (d Dataset) isSplittable() bool {

	// the trainingset and the testset should have the same datafile id
	if d.TrainingDataset.DatafileID != d.TestDataset.DatafileID {
		return false
	}

	// the split percentages should both be non-zero
	if int(d.TrainingDataset.SplitPercentage) == 0 || int(d.TestDataset.SplitPercentage) == 0 {
		return false
	}

	return true

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
