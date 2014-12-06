package elasticthought

import (
	"fmt"

	"github.com/couchbaselabs/logg"
)

// Worker job that downloads a datafile url contents to cbfs
type DatafileDownloader struct {
	Configuration Configuration
	Datafile      Datafile
}

// Run this job
func (d DatafileDownloader) Run() {

	logg.LogTo("DATAFILE_DOWNLOADER", "datafile downloader run()")

	db := d.Configuration.DbConnection()

	// create a new cbfs client
	cbfs, err := d.Configuration.NewCbfsClient()
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		d.recordProcessingError(errMsg)
		return
	}

	// copy url contents to cbfs
	logg.LogTo("DATAFILE_DOWNLOADER", "copytocbfs: %+v %v %v", d, db, cbfs)
	cbfsUrl, err := d.Datafile.CopyToCBFS(db, cbfs)
	if err != nil {
		d.recordProcessingError(err)
		return
	}

	// Update the state of the dataset to be finished
	d.Datafile.Url = cbfsUrl
	if err := d.Datafile.FinishedSuccessfully(db); err != nil {
		errMsg := fmt.Errorf("Error marking datafile %+v finished: %v", d, err)
		d.recordProcessingError(errMsg)
		return
	}

}

// Codereview: de-dupe -- dataset_splitter has same method
func (d DatafileDownloader) recordProcessingError(err error) {
	logg.LogError(err)
	db := d.Configuration.DbConnection()
	if err := d.Datafile.Failed(db, err); err != nil {
		errMsg := fmt.Errorf("Error setting datafile as failed: %v", err)
		logg.LogError(errMsg)
	}
}
