package elasticthought

import (
	"archive/tar"
	"bytes"
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
	tr, err := d.openTarGzStream(datafile.Url)
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

// Opens to tar.gz streams
func (d DatasetSplitter) openTarGzStream(url string) (*tar.Reader, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(gzipReader)

	return tarReader, nil

}

/*
Read from source tar stream and write training and test to given tar writers

It is a bit complicated because the source stream should be iterated over _once_,
but it needs to workaround the inability to look ahead or behind in the stream.
(You can't know how to split the files within a directory until you've seen
all the files in that directory).

What it does is:

    - Iterate over the reader
    - Keeps writing entries to a "temp tar writer" as long as we are in the same dir
    - Once we hit a new dir (or the end of the tar):
        - Look at all files in temp tar writer and calculate target distribution
        - Iterate over temp tar writer and write to target tar writers
        - Reset tar writer

The temp tar writer mentioned above currently saves everything to a bytes.Buffer,
but if needed it can easily switched over to use a

*/
func (d DatasetSplitter) transform(source *tar.Reader, train, test *tar.Writer) error {

	currentDirectory := ""
	currentDirFiles := []string{}

	buf := new(bytes.Buffer)
	twTemp := tar.NewWriter(buf)
	defer twTemp.Close()

	for {
		hdr, err := source.Next()

		if err == io.EOF {
			// end of tar archive - split up the saved buffer and distribute to tar writers
			twTemp.Close()
			trTemp := tar.NewReader(buf)
			if err := d.split(trTemp, train, test, currentDirFiles); err != nil {
				return err
			}
			break
		}

		if err != nil {
			return err
		}

		// figure out directory name of this file
		pathComponents := strings.Split(hdr.Name, "/")
		if len(pathComponents) != 2 {
			return fmt.Errorf("Path does not have 2 components: %v", hdr.Name)
		}
		directory := pathComponents[0]

		// its the first directory we've seen, set to current
		if currentDirectory == "" {
			currentDirectory = directory
		}

		// function which copies an entry over to temp tar writer and
		// adds the current file to the running list of files for this dir
		addToTwTemp := func() error {
			if err := twTemp.WriteHeader(hdr); err != nil {
				return err
			}
			_, err = io.Copy(twTemp, source)
			if err != nil {
				return err
			}
			// save this file to the list of files we've accumulated for this dir
			currentDirFiles = append(currentDirFiles, hdr.Name)
			return nil
		}

		switch directory {
		case currentDirectory:
			// we're in the same directory

			// add to temp tar writer
			if err := addToTwTemp(); err != nil {
				return err
			}

		default:

			// we're in a new directory
			currentDirectory = directory

			// split up the saved buffer and distribute to tar writers
			twTemp.Close()
			trTemp := tar.NewReader(buf)
			if err := d.split(trTemp, train, test, currentDirFiles); err != nil {
				return err
			}

			// reset the tar writer to create a fresh one
			twTemp.Close()
			buf = new(bytes.Buffer)
			twTemp = tar.NewWriter(buf)

			// reset file list
			currentDirFiles = []string{}

			// add to temp tar writer
			if err := addToTwTemp(); err != nil {
				return err
			}

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

func (d DatasetSplitter) split(source *tar.Reader, train, test *tar.Writer, files []string) error {

	numTraining := int(float64(len(files)) * d.Dataset.TrainingDataset.SplitPercentage)

	numTest := len(files) - int(numTraining)

	// split files into subsets based on ratios in dataset
	trainingFiles, testFiles, err := splitFilesToMaps(files, numTraining, numTest)

	if err != nil {
		return err
	}

	for {
		hdr, err := source.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return err
		}

		// figure out which tar writer to use
		var tw *tar.Writer
		if _, ok := trainingFiles[hdr.Name]; ok {
			tw = train
		} else if _, ok := testFiles[hdr.Name]; ok {
			tw = test
		} else {
			return fmt.Errorf("file not in either set: %v", hdr.Name)
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		_, err = io.Copy(tw, source)
		if err != nil {
			return err
		}

	}

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

func splitFilesToMaps(files []string, numTraining, numTest int) (training map[string]struct{}, test map[string]struct{}, err error) {

	training = make(map[string]struct{})
	test = make(map[string]struct{})

	for i, file := range files {
		if i < numTraining {
			training[file] = struct{}{}
		} else {
			test[file] = struct{}{}
		}
	}
	return
}
