package elasticthought

import "github.com/couchbaselabs/logg"

// Worker job that splits a dataset into training/test set
type DatasetSplitter struct {
	Configuration Configuration
	Dataset       Dataset
}

func (d DatasetSplitter) Run() {
	logg.LogTo("DATASET_SPLITTER", "Datasetsplitter.run()!.  Config: %+v Dataset: %+v", d.Configuration, d.Dataset)
}
