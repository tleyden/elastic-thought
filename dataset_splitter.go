package elasticthought

import (
	"archive/tar"
	"fmt"
	"io"
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

// Read from source tar stream and write training and test to given tar writers
func (d DatasetSplitter) transform(source *tar.Reader, train, test *tar.Writer) error {

	// build a map from the source

	// split the map into training and test

	// iterate over the source

	// distribute to writers based on training and test maps

	return nil
}
