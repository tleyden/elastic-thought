package elasticthought

import (
	"io"

	"github.com/couchbaselabs/logg"
)

// Worker job that splits a dataset into training/test set
type DatasetSplitter struct {
	Configuration Configuration
	Dataset       Dataset
}

func (d DatasetSplitter) Run() {

	logg.LogTo("DATASET_SPLITTER", "Datasetsplitter.run()!.  Config: %+v Dataset: %+v", d.Configuration, d.Dataset)

	// Find the datafile object associated with dataset

	// Get the url associated with datafile

	// Open the url -- content type should be application/x-gzip and url should end with
	// .tar.gz

	// Read from the stream and open tar archive

	// Walk the directories and split the files

	// Write to training and test tar archive

	// Save both training and test tar archive to cbfs (wrapped in gzip stream)

}

// Read from source and write training and test to given writers
func (d DatasetSplitter) transform(source io.Reader, train, test io.Writer) error {
	return nil
}
