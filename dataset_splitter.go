package elasticthought

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
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

// TODO: this is currently reading the entire tar stream into a buffer before
// handing it to cbfs.  Should be using an io.Pipe.
func (d DatasetSplitter) Run() {

	logg.LogTo("DATASET_SPLITTER", "Datasetsplitter.run()!.  Config: %+v Dataset: %+v", d.Configuration, d.Dataset)

	// Find the datafile object associated with dataset
	db := d.Configuration.DbConnection()
	datafile, err := d.Dataset.GetDatafile(db)
	if err != nil {
		errMsg := fmt.Errorf("Error looking up datafile with id: %v.  Error: %v", d.Dataset.DatafileID, err)
		logg.LogError(errMsg)
		return
	}

	// Open the url -- content type should be application/x-gzip
	tr1, tr2, err := d.openTwoTarGzStreams(datafile.Url)
	if err != nil {
		errMsg := fmt.Errorf("Error opening tar.gz streams: %v", err)
		logg.LogError(errMsg)
		return
	}

	bufferTesting := &bytes.Buffer{}
	bufferTraining := &bytes.Buffer{}

	tarWriterTesting := tar.NewWriter(bufferTesting)
	tarWriterTraining := tar.NewWriter(bufferTraining)

	cbfs, err := cbfsclient.New(d.Configuration.CbfsUrl)
	logg.LogTo("DATASET_SPLITTER", "Created cbfs client: %v", cbfs)
	options := cbfsclient.PutOptions{
		ContentType: "application/x-gzip",
	}
	logg.LogTo("DATASET_SPLITTER", "options: %v", options)

	destTraining := fmt.Sprintf("%v/training.tar.gz", d.Dataset.Id)
	destTesting := fmt.Sprintf("%v/testing.tar.gz", d.Dataset.Id)
	logg.LogTo("DATASET_SPLITTER", " %v %v ", destTesting, destTraining)

	logg.LogTo("DATASET_SPLITTER", "Calling transform")
	err = d.transform(tr1, tr2, tarWriterTraining, tarWriterTesting)
	if err != nil {
		errMsg := fmt.Errorf("Error transforming tar stream: %v", err)
		logg.LogError(errMsg)
		return
	}

	// close writers
	logg.LogTo("DATASET_SPLITTER", "Closing writers")
	if err := tarWriterTesting.Close(); err != nil {
		errMsg := fmt.Errorf("Error closing tar writer: %v", err)
		logg.LogError(errMsg)
		return
	}
	if err := tarWriterTraining.Close(); err != nil {
		errMsg := fmt.Errorf("Error closing tar reader: %v", err)
		logg.LogError(errMsg)
		return
	}
	logg.LogTo("DATASET_SPLITTER", "Closed writers")

	// At this point bufferTesting has all data
	logg.LogTo("DATASET_SPLITTER", "bufferTesting size: %d", bufferTesting.Len())
	logg.LogTo("DATASET_SPLITTER", "bufferTraining size: %d", bufferTraining.Len())

	if err := cbfs.Put("", destTesting, bufferTesting, options); err != nil {
		errMsg := fmt.Errorf("Error writing %v to cbfs: %v", destTesting, err)
		logg.LogError(errMsg)
		return

	}
	logg.LogTo("DATASET_SPLITTER", "Wrote %v to cbfs", destTesting)

	if err := cbfs.Put("", destTraining, bufferTraining, options); err != nil {
		errMsg := fmt.Errorf("Error writing %v to cbfs: %v", destTraining, err)
		logg.LogError(errMsg)
		return
	}
	logg.LogTo("DATASET_SPLITTER", "Wrote %v to cbfs", destTraining)

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

		// TODO: is there a more efficient way to do this?
		bytes, err := ioutil.ReadAll(source2)
		if err != nil {
			return err
		}

		if err := twToAdd.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := twToAdd.Write(bytes); err != nil {
			return err
		}

	}

	logg.LogTo("DATASET_SPLITTER", "done iterating over source")

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
