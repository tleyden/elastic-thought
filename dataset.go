package elasticthought

import (
	"fmt"

	"github.com/couchbaselabs/logg"
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
	ProcessingLog   string          `json:"processing-log"`
	TrainingDataset TrainingDataset `json:"training" binding:"required"`
	TestDataset     TestDataset     `json:"test" binding:"required"`

	// had to make exported, due to https://github.com/gin-gonic/gin/pull/123
	// waiting for this to get merged into master branch, since go get
	// pulls from master branch.
	Configuration Configuration
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
func NewDataset(c Configuration) *Dataset {
	return &Dataset{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_DATASET},
		Configuration:     c,
	}
}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (d *Dataset) Insert() error {

	db := d.Configuration.DbConnection()

	id, rev, err := db.Insert(d)
	if err != nil {
		err := fmt.Errorf("Error inserting: %+v.  Err: %v", d, err)
		return err
	}

	d.Id = id
	d.Revision = rev

	return nil

}

// Find and return the datafile associated with this dataset
func (d Dataset) GetSplittableDatafile(db couch.Database) (*Datafile, error) {

	if !d.isSplittable() {
		return nil, fmt.Errorf("This dataset is not splittable")
	}

	logg.LogTo("DATASET_SPLITTER", "Looking up data file: %v", d.TrainingDataset.DatafileID)

	// choose either datafile id since they are the same
	return FindDatafile(db, d.TrainingDataset.DatafileID)

}

// Get the training datafile object
func (d Dataset) GetTrainingDatafile(db couch.Database) (*Datafile, error) {
	return FindDatafile(db, d.TrainingDataset.DatafileID)
}

// Get the testing datafile object
func (d Dataset) GetTestingDatafile(db couch.Database) (*Datafile, error) {
	return FindDatafile(db, d.TestDataset.DatafileID)
}

// Get the source url associated with the training datafile
func (d Dataset) GetTrainingDatafileUrl(db couch.Database) string {
	datafile, err := d.GetTrainingDatafile(db)
	if err != nil {
		return fmt.Sprintf("error getting training datafile url: %v", err)
	}
	return datafile.Url
}

// Get the source url associated with the testing datafile
func (d Dataset) GetTestingDatafileUrl(db couch.Database) string {
	datafile, err := d.GetTestingDatafile(db)
	if err != nil {
		return fmt.Sprintf("error getting testing datafile url: %v", err)
	}
	return datafile.Url
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

// Update the processing state to new state.
func (d *Dataset) UpdateProcessingState(newState ProcessingState) (bool, error) {

	updater := func(dataset *Dataset) {
		dataset.ProcessingState = newState
	}

	doneMetric := func(dataset Dataset) bool {
		return dataset.ProcessingState == newState
	}

	return d.casUpdate(updater, doneMetric)

}

func (d *Dataset) casUpdate(updater func(*Dataset), doneMetric func(Dataset) bool) (bool, error) {

	db := d.Configuration.DbConnection()

	genUpdater := func(datasetPtr interface{}) {
		cjp := datasetPtr.(*Dataset)
		updater(cjp)
	}

	genDoneMetric := func(datasetPtr interface{}) bool {
		cjp := datasetPtr.(*Dataset)
		return doneMetric(*cjp)
	}

	refresh := func(datasetPtr interface{}) error {
		cjp := datasetPtr.(*Dataset)
		return cjp.RefreshFromDB(db)
	}

	return casUpdate(db, d, genUpdater, genDoneMetric, refresh)

}

// Update the dataset state to record that it finished successfully
// Codereview: de-dupe with datafile FinishedSuccessfully
func (d Dataset) FinishedSuccessfully(db couch.Database) error {

	_, err := d.UpdateProcessingState(FinishedSuccessfully)
	if err != nil {
		return err
	}

	return nil

}

// Update the dataset state to record that it failed
// Codereview: datafile.go has same method
func (d Dataset) Failed(db couch.Database, processingErr error) error {

	_, err := d.UpdateProcessingState(Failed)
	if err != nil {
		return err
	}

	processingLog := fmt.Sprintf("%v", processingErr)
	_, err = d.UpdateProcessingLog(processingLog)
	if err != nil {
		return err
	}

	return nil

}

func (d *Dataset) UpdateProcessingLog(val string) (bool, error) {

	updater := func(dataset *Dataset) {
		dataset.ProcessingLog = val
	}

	doneMetric := func(dataset Dataset) bool {
		return dataset.ProcessingLog == val
	}

	return d.casUpdate(updater, doneMetric)

}

func (d *Dataset) UpdateArtifactUrls(trainingDatasetUrl, testingDatasetUrl string) (bool, error) {

	updater := func(dataset *Dataset) {
		dataset.TrainingDataset.Url = trainingDatasetUrl
		dataset.TestDataset.Url = testingDatasetUrl
	}

	doneMetric := func(dataset Dataset) bool {
		return (dataset.TrainingDataset.Url == trainingDatasetUrl &&
			dataset.TestDataset.Url == testingDatasetUrl)
	}

	return d.casUpdate(updater, doneMetric)

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
func (d *Dataset) AddArtifactUrls() error {

	trainingDatasetUrl := fmt.Sprintf("%v%v", CBFS_URI_PREFIX, d.TrainingArtifactPath())
	testingDatasetUrl := fmt.Sprintf("%v%v", CBFS_URI_PREFIX, d.TestingArtifactPath())

	_, err := d.UpdateArtifactUrls(trainingDatasetUrl, testingDatasetUrl)
	return err

}

func (d *Dataset) GetProcessingState() ProcessingState {
	return d.ProcessingState
}

func (d *Dataset) SetProcessingState(newState ProcessingState) {
	d.ProcessingState = newState
}

func (d *Dataset) RefreshFromDB(db couch.Database) error {
	dataset := Dataset{}
	err := db.Retrieve(d.Id, &dataset)
	if err != nil {
		logg.LogTo("TRAINING_JOB", "Error getting latest: %v", err)
		return err
	}
	*d = dataset
	return nil
}
