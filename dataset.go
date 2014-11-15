package elasticthought

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
}

type TrainingDataset struct {
	SplitPercentage float64 `json:"split-percentage"`
}

type TestDataset struct {
	SplitPercentage float64 `json:"split-percentage"`
}

// Create a new dataset
func NewDataset() *Dataset {
	return &Dataset{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_DATASET},
	}
}
