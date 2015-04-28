package elasticthought

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/couchbaselabs/logg"
)

// Worker job that splits a dataset into training/test set
type DatasetSplitter struct {
	Configuration Configuration
	Dataset       Dataset
}

// Run this job
func (d DatasetSplitter) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	dataset := &d.Dataset

	updatedState, err := dataset.UpdateProcessingState(Processing)
	if err != nil {
		d.recordProcessingError(err)
		return
	}
	if !updatedState {
		logg.LogTo("TRAINING_JOB", "%+v already processed.  Ignoring.", d)
		return
	}

	switch d.Dataset.isSplittable() {
	case true:
		d.SplitDatafile()
	default:
		d.DownloadDatafiles()
	}

}

func (d DatasetSplitter) SplitDatafile() {

	// Find the datafile object associated with dataset
	db := d.Configuration.DbConnection()
	datafile, err := d.Dataset.GetSplittableDatafile(db)
	if err != nil {
		errMsg := fmt.Errorf("Error looking up datafile: %+v.  Error: %v", d.Dataset, err)
		d.recordProcessingError(errMsg)
		return
	}

	// Open the url -- content type should be application/x-gzip
	tr, err := openTarGzStream(datafile.Url)
	if err != nil {
		errMsg := fmt.Errorf("Error opening tar.gz streams: %v", err)
		d.recordProcessingError(errMsg)
		return
	}

	// Create pipes
	prTrain, pwTrain := io.Pipe()
	prTest, pwTest := io.Pipe()

	// Wrap in gzip writers
	pwGzTest := gzip.NewWriter(pwTest)
	pwGzTrain := gzip.NewWriter(pwTrain)

	// Create tar writers on the write end of the pipes
	tarWriterTesting := tar.NewWriter(pwGzTest)
	tarWriterTraining := tar.NewWriter(pwGzTrain)

	// Create a cbfs client
	cbfs, err := NewBlobStore(d.Configuration.CbfsUrl)
	options := BlobPutOptions{}
	options.ContentType = "application/x-gzip"
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		d.recordProcessingError(errMsg)
		return
	}

	// Figure out where to store these on cbfs
	destTraining := d.Dataset.TrainingArtifactPath()
	destTesting := d.Dataset.TestingArtifactPath()

	// Spawn a goroutine that will read from tar.gz reader coming from url data
	// and write to the training and test tar writers (which are on write ends of pipe)
	transformDoneChan := make(chan error, 1)
	go func() {

		// Must close _underlying_ piped writers, or the piped readers will
		// never get an EOF.  Closing just the tar writers that wrap the underlying
		// piped writers is not enough.
		defer pwTest.Close()
		defer pwTrain.Close()
		defer pwGzTest.Close()
		defer pwGzTrain.Close()

		logg.LogTo("DATASET_SPLITTER", "Calling transform")
		err = d.transform(tr, tarWriterTraining, tarWriterTesting)
		if err != nil {
			errMsg := fmt.Errorf("Error transforming tar stream: %v", err)
			logg.LogError(errMsg)
			transformDoneChan <- errMsg
			return
		}

		transformDoneChan <- nil

	}()

	// Spawn goroutines to read off the read ends of the pipe and store in cbfs
	cbfsTrainDoneChan := make(chan error, 1)
	cbfsTestDoneChan := make(chan error, 1)
	go func() {
		if err := cbfs.Put("", destTesting, prTest, options); err != nil {
			errMsg := fmt.Errorf("Error writing %v to cbfs: %v", destTesting, err)
			logg.LogError(errMsg)
			cbfsTestDoneChan <- errMsg
			return

		}
		logg.LogTo("DATASET_SPLITTER", "Wrote %v to cbfs", destTesting)
		cbfsTestDoneChan <- nil
	}()
	go func() {
		if err := cbfs.Put("", destTraining, prTrain, options); err != nil {
			errMsg := fmt.Errorf("Error writing %v to cbfs: %v", destTraining, err)
			logg.LogError(errMsg)
			cbfsTrainDoneChan <- errMsg
			return
		}
		logg.LogTo("DATASET_SPLITTER", "Wrote %v to cbfs", destTraining)
		cbfsTrainDoneChan <- nil
	}()

	// Wait for the results from all the goroutines
	cbfsTrainResult := <-cbfsTrainDoneChan
	cbfsTestResult := <-cbfsTestDoneChan
	transformResult := <-transformDoneChan

	// If any results had an error, log it and return
	results := []error{transformResult, cbfsTestResult, cbfsTrainResult}
	for _, result := range results {
		if result != nil {
			logg.LogTo("DATASET_SPLITTER", "Setting dataset to failed: %v", result)
			d.Dataset.Failed(db, fmt.Errorf("%v", result))
			return
		}
	}

	// Update the state of the dataset to be finished
	d.Dataset.FinishedSuccessfully(db)
	if err := d.Dataset.FinishedSuccessfully(db); err != nil {
		errMsg := fmt.Errorf("Error marking dataset %+v finished: %v", d, err)
		d.recordProcessingError(errMsg)
		return
	}

}

func (d DatasetSplitter) DownloadDatafiles() {

	// Create a cbfs client
	cbfs, err := NewBlobStore(d.Configuration.CbfsUrl)
	options := BlobPutOptions{}
	options.ContentType = "application/x-gzip"
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		d.recordProcessingError(errMsg)
		return
	}

	db := d.Configuration.DbConnection()

	source2destEntries := []struct {
		Url      string
		DestPath string
	}{
		{
			Url:      d.Dataset.GetTrainingDatafileUrl(db),
			DestPath: d.Dataset.TrainingArtifactPath(),
		},
		{
			Url:      d.Dataset.GetTestingDatafileUrl(db),
			DestPath: d.Dataset.TestingArtifactPath(),
		},
	}

	for _, source2destEntry := range source2destEntries {

		// open tar.gz stream to source
		// Open the url -- content type should be application/x-gzip

		func() {

			resp, err := http.Get(source2destEntry.Url)

			if err != nil {
				errMsg := fmt.Errorf("Error opening stream to: %v. Err %v", source2destEntry.Url, err)
				d.recordProcessingError(errMsg)
				return
			}
			defer resp.Body.Close()

			if err := cbfs.Put("", source2destEntry.DestPath, resp.Body, options); err != nil {
				errMsg := fmt.Errorf("Error writing %v to cbfs: %v", source2destEntry.DestPath, err)
				d.recordProcessingError(errMsg)
				return
			}

			logg.LogTo("DATASET_SPLITTER", "Wrote %v to cbfs", source2destEntry.DestPath)

		}()

	}

	// Update the state of the dataset to be finished
	if err := d.Dataset.FinishedSuccessfully(db); err != nil {
		errMsg := fmt.Errorf("Error marking dataset %+v finished: %v", d, err)
		d.recordProcessingError(errMsg)
		return
	}

}

// Codereview: de-dupe
func (d DatasetSplitter) recordProcessingError(err error) {
	logg.LogError(err)
	db := d.Configuration.DbConnection()
	if err := d.Dataset.Failed(db, err); err != nil {
		errMsg := fmt.Errorf("Error setting dataset as failed: %v", err)
		logg.LogError(errMsg)
	}
}

// Read from source tar stream and write training and test to given tar writers
func (d DatasetSplitter) transform(source *tar.Reader, train, test *tar.Writer) error {

	splitter := d.splitter(train, test)

	for {
		hdr, err := source.Next()

		if err == io.EOF {
			// end of tar archive
			break
		}

		if err != nil {
			return err
		}

		tw := splitter(hdr.Name)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = io.Copy(tw, source)
		if err != nil {
			return err
		}

	}

	// close writers
	if err := train.Close(); err != nil {
		errMsg := fmt.Errorf("Error closing tar writer: %v", err)
		logg.LogError(errMsg)
		return err
	}
	if err := test.Close(); err != nil {
		errMsg := fmt.Errorf("Error closing tar reader: %v", err)
		logg.LogError(errMsg)
		return err
	}

	return nil
}

func (d DatasetSplitter) splitter(train, test *tar.Writer) func(string) *tar.Writer {

	trainingRatio := int(d.Dataset.TrainingDataset.SplitPercentage * 100)
	testRatio := int(d.Dataset.TestDataset.SplitPercentage * 100)

	ratio := [2]int{trainingRatio, testRatio}

	dircounts := make(map[string][2]int)
	return func(file string) *tar.Writer {
		dir := path.Dir(file)
		counts := dircounts[dir]
		count0ratio1 := counts[0] * ratio[1]
		count1ratio0 := counts[1] * ratio[0]
		if count0ratio1 <= count1ratio0 {
			counts[0]++
			dircounts[dir] = counts
			return train
		} else {
			counts[1]++
			dircounts[dir] = counts
			return test
		}
	}
}

// Validate that the source tar stream conforms to expected specs
func (d DatasetSplitter) validate(source *tar.Reader) (bool, error) {

	// validation rules:
	// 1. has at least 2 files
	// 2. the depth of each file is 2 (folder/filename.xxx)

	numFiles := 0
	for {
		hdr, err := source.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return false, err
		}
		numFiles += 1

		pathComponents := strings.Split(hdr.Name, "/")
		if len(pathComponents) != 2 {
			return false, fmt.Errorf("Path does not have 2 components: %v", hdr.Name)
		}

	}

	if numFiles < 2 {
		return false, fmt.Errorf("Archive must contain at least 2 files")
	}

	return true, nil
}
