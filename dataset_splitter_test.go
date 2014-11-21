package elasticthought

import (
	"archive/tar"
	"bytes"
	"io"
	"log"
	"path"
	"testing"

	"github.com/couchbaselabs/go.assert"
)

type tarFile struct {
	Name string
	Body string
}

func init() {
	EnableAllLogKeys()
}

func create5050Splitter() DatasetSplitter {
	dataset := Dataset{
		TrainingDataset: TrainingDataset{
			SplitPercentage: 0.5,
		},
		TestDataset: TestDataset{
			SplitPercentage: 0.5,
		},
	}

	splitter := DatasetSplitter{
		Dataset: dataset,
	}
	return splitter
}

func create8020Splitter() DatasetSplitter {
	dataset := Dataset{
		TrainingDataset: TrainingDataset{
			SplitPercentage: 0.8,
		},
		TestDataset: TestDataset{
			SplitPercentage: 0.2,
		},
	}

	splitter := DatasetSplitter{
		Dataset: dataset,
	}
	return splitter
}

func TestTransform(t *testing.T) {

	splitter := create8020Splitter()

	// Create a test tar archive
	buf := new(bytes.Buffer)

	var files = []tarFile{
		{"foo/1.txt", "."},
		{"foo/2.txt", "."},
		{"bar/1.txt", "."},
		{"bar/2.txt", "."},
		{"bar/3.txt", "."},
		{"bar/4.txt", "."},
		{"bar/5.txt", "."},
	}
	createArchive(buf, files)

	// Open the tar archive for reading.
	r := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(r)

	// create two writers
	bufTrain := new(bytes.Buffer)
	bufTest := new(bytes.Buffer)
	twTrain := tar.NewWriter(bufTrain)
	twTest := tar.NewWriter(bufTest)

	// pass these into transform
	err := splitter.transform2(tr, twTrain, twTest)
	assert.True(t, err == nil)
	if err != nil {
		assert.Errorf(t, "Err from transform2: %v", err)
	}

	trainingResult := make(filemap)
	testResult := make(filemap)

	// assert that the both the training and test tar archives
	// have been split correctly over each "label" folder
	// Also, the resulting split sets must be disjoint.
	buffers := []*bytes.Buffer{bufTrain, bufTest}
	for _, buffer := range buffers {
		readerVerify := bytes.NewReader(buffer.Bytes())
		trVerify := tar.NewReader(readerVerify)

		for {
			hdr, err := trVerify.Next()
			if err == io.EOF {
				// end of tar archive
				break
			}
			if err != nil {
				log.Fatalln(err)
			}
			assert.Equals(t, hdr.Uid, 100)
			assert.Equals(t, hdr.Gid, 101)

			dir := path.Dir(hdr.Name)

			switch buffer {
			case bufTrain:
				trainingResult.addFileToDirectory(dir, hdr.Name)
			case bufTest:
				testResult.addFileToDirectory(dir, hdr.Name)
			}

		}

	}

	// make sure they have correct number of entries
	assert.Equals(t, len(trainingResult["foo"]), 1)
	assert.Equals(t, len(testResult["foo"]), 1)
	assert.Equals(t, len(trainingResult["bar"]), 4)
	assert.Equals(t, len(testResult["foo"]), 1)

}

func TestValidateValid(t *testing.T) {

	buf := new(bytes.Buffer)
	var files = []tarFile{
		{"foo/1.txt", "Hello 1."},
		{"foo/2.txt", "Hello 2."},
		{"bar/1.txt", "Hello bar 1."},
		{"bar/2.txt", "Hello bar 2."},
	}
	createArchive(buf, files)
	reader := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(reader)

	splitter := DatasetSplitter{}
	ok, err := splitter.validate(tr)
	assert.True(t, ok)
	assert.True(t, err == nil)

}

func TestValidateTooDeep(t *testing.T) {

	buf := new(bytes.Buffer)
	var files = []tarFile{
		{"a/foo/1.txt", "Hello 1."},
		{"a/bar/1.txt", "Hello bar 1."},
	}
	createArchive(buf, files)
	reader := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(reader)

	splitter := DatasetSplitter{}
	ok, err := splitter.validate(tr)
	assert.False(t, ok)
	assert.True(t, err != nil)

}

func createArchive(buf *bytes.Buffer, tarFiles []tarFile) {

	// Create a new tar archive.
	tw := tar.NewWriter(buf)

	for _, file := range tarFiles {
		hdr := &tar.Header{
			Name: file.Name,
			Size: int64(len(file.Body)),
			Uid:  100,
			Gid:  101,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
		}
	}
	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		log.Fatalln(err)
	}

}
