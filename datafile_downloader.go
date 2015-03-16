package elasticthought

import (
	"fmt"
	"sync"

	"github.com/couchbaselabs/logg"
)

// Worker job that downloads a datafile url contents to cbfs
type DatafileDownloader struct {
	Configuration Configuration
	Datafile      Datafile
}

// Run this job
func (d DatafileDownloader) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	logg.LogTo("DATAFILE_DOWNLOADER", "datafile downloader run()")

	db := d.Configuration.DbConnection()
	datafile := &d.Datafile
	updatedState, err := datafile.UpdateProcessingState(Processing)
	if err != nil {
		d.recordProcessingError(err)
		return
	}
	if !updatedState {
		logg.LogTo("DATAFILE_DOWNLOADER", "%+v already processed.  Ignoring.", d)
		return
	}

	// create a new cbfs client
	cbfs, err := d.Configuration.NewBlobStoreClient()
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		d.recordProcessingError(errMsg)
		return
	}

	// copy url contents to cbfs
	logg.LogTo("DATAFILE_DOWNLOADER", "Put to CBFS: %+v %v %v", d, db, cbfs)
	cbfsDestPath, err := d.Datafile.CopyToBlobStore(db, cbfs)
	if err != nil {
		d.recordProcessingError(err)
		return
	}

	// build a url to the cbfs file
	cbfsUrl := fmt.Sprintf("%v/%v", d.Configuration.CbfsUrl, cbfsDestPath)
	d.Datafile.Url = cbfsUrl

	// Update the state of the dataset to be finished
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
