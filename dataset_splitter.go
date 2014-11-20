package elasticthought

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/couchbaselabs/cbfs/client"
	"github.com/couchbaselabs/logg"
)

// Worker job that splits a dataset into training/test set
type DatasetSplitter struct {
	Configuration Configuration
	Dataset       Dataset
}

// Run this job
func (d DatasetSplitter) Run() {

	// Find the datafile object associated with dataset
	db := d.Configuration.DbConnection()
	datafile, err := d.Dataset.GetDatafile(db)
	if err != nil {
		errMsg := fmt.Errorf("Error looking up datafile with id: %v.  Error: %v", d.Dataset.DatafileID, err)
		d.recordProcessingError(errMsg)
		return
	}

	// Open the url -- content type should be application/x-gzip
	tr1, tr2, err := d.openTwoTarGzStreams(datafile.Url)
	if err != nil {
		errMsg := fmt.Errorf("Error opening tar.gz streams: %v", err)
		d.recordProcessingError(errMsg)
		return
	}

	// Create pipes
	prTrain, pwTrain := io.Pipe()
	prTest, pwTest := io.Pipe()

	// Create tar writers on the write end of the pipes
	tarWriterTesting := tar.NewWriter(pwTest)
	tarWriterTraining := tar.NewWriter(pwTrain)

	// Create a cbfs client
	cbfs, err := cbfsclient.New(d.Configuration.CbfsUrl)
	options := cbfsclient.PutOptions{
		ContentType: "application/x-gzip",
	}
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

		logg.LogTo("DATASET_SPLITTER", "Calling transform")
		err = d.transform(tr1, tr2, tarWriterTraining, tarWriterTesting)
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
			logg.LogTo("DATASET_SPLITTER", "Setting dataset to failed: %v", transformResult)
			d.Dataset.Failed(db, fmt.Errorf("%v", transformResult))
			return
		}
	}

	// Update the state of the dataset to be finished
	d.Dataset.FinishedSuccessfully(db)

}

func (d DatasetSplitter) recordProcessingError(err error) {
	logg.LogError(err)
	db := d.Configuration.DbConnection()
	if err := d.Dataset.Failed(db, err); err != nil {
		errMsg := fmt.Errorf("Error setting dataset as failed: %v", err)
		logg.LogError(errMsg)
	}
}

// Opens to tar.gz streams to the same url.  The reason this is done twice is due to
// ugly hack, which is documented in the transform() method.
func (d DatasetSplitter) openTwoTarGzStreams(url string) (*tar.Reader, *tar.Reader, error) {

	resp1, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	gzipReader1, err := gzip.NewReader(resp1.Body)
	if err != nil {
		return nil, nil, err
	}
	tarReader1 := tar.NewReader(gzipReader1)

	resp2, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	gzipReader2, err := gzip.NewReader(resp2.Body)
	if err != nil {
		return nil, nil, err
	}
	tarReader2 := tar.NewReader(gzipReader2)

	return tarReader1, tarReader2, nil

}

// Read from source tar stream and write training and test to given tar writers
//
// TODO: fix ugly hack.  Since I'm trying to read from the source stream *twice*, which
// doesn't work, the workaround is to expect *two* source streams: source1 and source2.
// That way after source1 is read, source2 is ready for reading from the beginning
func (d DatasetSplitter) transform(source1, source2 *tar.Reader, train, test *tar.Writer) error {

	// build a map from the source
	sourceMap, err := d.createMap(source1)
	if err != nil {
		return err
	}
	logg.Log("sourceMap: %+v", sourceMap)

	// split the map into training and test
	trainMap, testMap, err := d.splitMap(sourceMap)
	if err != nil {
		return err
	}

	// iterate over the source
	logg.LogTo("DATASET_SPLITTER", "iterate over source")
	for {
		hdr, err := source2.Next()

		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return err
		}

		// distribute to writers based on training and test maps
		var twToAdd *tar.Writer

		// if strings.HasPrefix(hdr.Name, "foo") {
		if trainMap.hasPath(hdr.Name) {
			// add to training tar writer
			twToAdd = train
		} else if testMap.hasPath(hdr.Name) {
			// add to testing tar writer
			twToAdd = test
		} else {
			logg.LogPanic("File not present in either test/train: %v", hdr.Name)
		}

		if err := twToAdd.WriteHeader(hdr); err != nil {
			return err
		}

		_, err = io.Copy(twToAdd, source2)
		if err != nil {
			return err
		}

	}

	logg.LogTo("DATASET_SPLITTER", "done iterating over source")

	// close writers
	logg.LogTo("DATASET_SPLITTER", "Closing writers")
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
	logg.LogTo("DATASET_SPLITTER", "Closed writers")

	return nil
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

// Create a map of folder -> []filename for all entries in the archive
func (d DatasetSplitter) createMap(source *tar.Reader) (filemap, error) {

	resultMap := filemap{}
	for {
		hdr, err := source.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return nil, err
		}

		pathComponents := strings.Split(hdr.Name, "/")

		if len(pathComponents) != 2 {
			return nil, fmt.Errorf("Path does not have 2 components: %v", hdr.Name)
		}

		directory := pathComponents[0]
		filename := pathComponents[1]

		resultMap.addFileToDirectory(directory, filename)

	}

	return resultMap, nil

}

// Split map into training and testing disjoint subsets based on values
// of DatasetSplitter's Dataset
func (d DatasetSplitter) splitMap(source filemap) (training filemap, testing filemap, err error) {

	training = filemap{}
	testing = filemap{}

	// iterate over source keys
	for directory, files := range source {

		numTraining := int(float64(len(files)) * d.Dataset.TrainingDataset.SplitPercentage)

		numTest := len(files) - int(numTraining)

		// split files into subsets based on ratios in dataset
		trainingFiles, testFiles, err := splitFiles(files, numTraining, numTest)

		if err != nil {
			return nil, nil, err
		}

		// add to respective maps
		training[directory] = trainingFiles
		testing[directory] = testFiles

	}

	return training, testing, nil

}

func splitFiles(files []string, numTraining, numTest int) (training []string, test []string, err error) {
	for i, file := range files {
		if i < numTraining {
			training = append(training, file)
		} else {
			test = append(test, file)
		}
	}
	return
}
