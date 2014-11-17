package elasticthought

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/couchbaselabs/logg"
)

// Worker job that splits a dataset into training/test set
type DatasetSplitter struct {
	Configuration Configuration
	Dataset       Dataset
}

type filemap map[string][]string

func (f filemap) addFileToDirectory(directory, fileToAdd string) {
	files, ok := f[directory]
	if !ok {
		files = []string{}
		f[directory] = files
	}
	files = append(files, fileToAdd)
	f[directory] = files
}

func (f filemap) hasPath(path string) bool {
	pathComponents := strings.Split(path, "/")
	directory := pathComponents[0]
	filename := pathComponents[1]
	files, ok := f[directory]
	if !ok {
		return false
	}
	for _, file := range files {
		if file == filename {
			return true
		}
	}
	return false
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

// Read from source tar stream and write training and test to given tar writers
func (d DatasetSplitter) transform(source *tar.Reader, train, test *tar.Writer) error {

	// build a map from the source
	sourceMap, err := d.createMap(source)
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
	logg.Log("iterate over source")
	for {
		hdr, err := source.Next()
		logg.Log("hdr: %v", hdr)

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
		bytes, err := ioutil.ReadAll(source)
		if err != nil {
			return err
		}

		logg.Log("file: %v numbytes: %v", hdr.Name, len(bytes))

		hdrToAdd := &tar.Header{
			Name: hdr.Name,
			Size: int64(len(bytes)),
		}
		if err := twToAdd.WriteHeader(hdrToAdd); err != nil {
			return err
		}
		if _, err := twToAdd.Write(bytes); err != nil {
			return err
		}

	}

	// close writers
	if err := train.Close(); err != nil {
		return err
	}
	if err := test.Close(); err != nil {
		return err
	}

	return nil
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
