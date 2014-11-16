package elasticthought

import "github.com/couchbaselabs/logg"

// Worker job that splits a dataset into training/test set
type DatasetSplitter struct {
	Dataset Dataset
}

func (d DatasetSplitter) Run() {
	logg.LogTo("DATASET_SPLITTER", "Datasetsplitter.run()!.  Dataset: %+v", d.Dataset)
}
