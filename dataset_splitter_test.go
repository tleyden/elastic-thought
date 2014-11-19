package elasticthought

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
)

type tarFile struct {
	Name string
	Body string
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

func TestTransform(t *testing.T) {

	splitter := create5050Splitter()

	// Create a test tar archive
	buf := new(bytes.Buffer)
	createTestArchive(buf)

	// Open the tar archive for reading.
	r := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(r)

	// ugly hack.  Since I'm trying to read from the source stream *twice*, which
	// doesn't work, the workaround is to create *two* source streams.  this code
	// creates the second source stream from the same buffer.
	r2 := bytes.NewReader(buf.Bytes())
	tr2 := tar.NewReader(r2)

	// create two writers
	bufTrain := new(bytes.Buffer)
	bufTest := new(bytes.Buffer)
	twTrain := tar.NewWriter(bufTrain)
	twTest := tar.NewWriter(bufTest)

	// pass these into transform
	err := splitter.transform(tr, tr2, twTrain, twTest)
	assert.True(t, err == nil)

	// assert that the both the training and test tar archives
	// have been split 50/50 over each "label" folder, so each "label"
	// folder should now have one example each.  Ie, one "foo/x.txt" and
	// one "bar/x.txt".  Also, the resulting split sets must be disjoint.
	buffers := []*bytes.Buffer{bufTrain, bufTest}
	for _, buffer := range buffers {
		readerVerify := bytes.NewReader(buffer.Bytes())
		trVerify := tar.NewReader(readerVerify)

		seenFoos := map[string]string{}
		seenBars := map[string]string{}
		numFoo := 0
		numBar := 0
		for {
			hdr, err := trVerify.Next()
			if err == io.EOF {
				// end of tar archive
				break
			}
			if err != nil {
				log.Fatalln(err)
			}
			logg.Log("filename: %v", hdr.Name)
			assert.Equals(t, hdr.Uid, 100)
			assert.Equals(t, hdr.Gid, 101)
			if strings.HasPrefix(hdr.Name, "foo") {
				numFoo += 1
				// make sure it's the first time seeing this filename
				if _, ok := seenFoos[hdr.Name]; ok {
					logg.LogPanic("Not first time seeing: %v", hdr.Name)
				}
				seenFoos[hdr.Name] = hdr.Name
			}
			if strings.HasPrefix(hdr.Name, "bar") {
				numBar += 1
				if _, ok := seenBars[hdr.Name]; ok {
					logg.LogPanic("Not first time seeing: %v", hdr.Name)
				}
				seenBars[hdr.Name] = hdr.Name

			}

		}

		assert.Equals(t, numFoo, 1)
		assert.Equals(t, numBar, 1)

	}

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

func TestCreateMap(t *testing.T) {

	buf := new(bytes.Buffer)
	var files = []tarFile{
		{"foo/1.txt", "Hello 1."},
		{"bar/1.txt", "Hello bar 1."},
		{"bar/2.txt", "Hello bar 2."},
	}
	createArchive(buf, files)
	reader := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(reader)

	splitter := DatasetSplitter{}
	filemap, err := splitter.createMap(tr)
	assert.True(t, err == nil)

	assert.Equals(t, len(filemap["foo"]), 1)
	assert.Equals(t, len(filemap["bar"]), 2)

}

func TestSplitMap(t *testing.T) {

	splitter := create5050Splitter()
	fmap := filemap{
		"foo": []string{"foo1.txt", "foo2.txt"},
		"bar": []string{"bar1.txt", "bar2.txt"},
	}
	train, test, err := splitter.splitMap(fmap)
	assert.True(t, err == nil)

	// assertions
	mapsToVerify := []filemap{train, test}
	for _, mapToVerify := range mapsToVerify {
		seenFoos := map[string]string{}
		seenBars := map[string]string{}
		numFoo := 0
		numBar := 0
		for _, files := range mapToVerify {
			for _, file := range files {
				if strings.Contains(file, "foo") {
					numFoo += 1
					// make sure it's the first time seeing this filename
					if _, ok := seenFoos[file]; ok {
						logg.LogPanic("Not first time seeing: %v", file)
					}
					seenFoos[file] = file
				}
				if strings.Contains(file, "bar") {
					numBar += 1
					// make sure it's the first time seeing this filename
					if _, ok := seenBars[file]; ok {
						logg.LogPanic("Not first time seeing: %v", file)
					}
					seenBars[file] = file
				}
			}
		}
		assert.Equals(t, numFoo, 1)
		assert.Equals(t, numBar, 1)
	}

}

func createTestArchive(buf *bytes.Buffer) {

	var files = []tarFile{
		{"foo/1.txt", "Hello 1."},
		{"foo/2.txt", "Hello 2."},
		{"bar/1.txt", "Hello bar 1."},
		{"bar/2.txt", "Hello bar 2."},
	}
	createArchive(buf, files)

}

// See https://groups.google.com/forum/#!topic/golang-nuts/l27h06d_2Gk
func TestSplitTar(t *testing.T) {

	// Create a test tar archive
	buf := new(bytes.Buffer)
	var files = []tarFile{
		{"foo/1.txt", "Hello 1."},
		{"foo/2.txt", "Hello 2."},
		{"foo/3.txt", "Hello 3."},
		{"foo/4.txt", "Hello 4."},
		{"foo/5.txt", "Hello 5."},
		{"bar/1.txt", "Hello bar 1."},
		{"bar/2.txt", "Hello bar 2."},
	}
	createArchive(buf, files)

	// Goal: split the tar file into two tar files, such that:
	// a percentage of the files in each directory go to the respective
	// tar files.  If it was a 50/50 split, it would look like:
	// tar1 -> [foo/1.txt, foo/2.txt, bar/1.txt]
	// tar2 -> [foo/3.txt, foo/4.txt, bar/2.txt]
	//
	// But other percentages should be supported, such as 80/20 split.

	// Create a tar reader
	tr := tar.NewReader(buf)

	// Create two tar writers for the destinations
	destBuf1 := new(bytes.Buffer)
	tw1 := tar.NewWriter(destBuf1)
	destBuf2 := new(bytes.Buffer)
	tw2 := tar.NewWriter(destBuf2)

	tars := []*tar.Writer{tw1, tw2}

	// fmt.Printf("%v %v %v %v", destBuf1, tw1, destBuf2, tw2)

	for i := 0; ; i++ {

		hdr, err := tr.Next()
		// fmt.Printf("%v", hdr)

		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			fmt.Printf("Got error, aborting.  %v", err)
			return
		}

		// In order to find where this entry should go, I need to
		// know how many entries there are in this directory.  (if its an
		// 80/20 split, and there are 10 files, then the first 8 files
		// should go into tw1 and the remaining 2 should go into tw2).
		// I can't read ahead, because there is no way to rewind the
		// stream.

		tw := tars[i%len(tars)]

		if err := tw.WriteHeader(hdr); err != nil {
			fmt.Printf("Got error, aborting.  %v", err)
			return
		}

		_, err = io.Copy(tw, tr)
		if err != nil {
			fmt.Printf("Got error, aborting.  %v", err)
			return
		}

	}

	tw1.Close()
	tw2.Close()

	fmt.Printf("looping\n")

	tr1 := tar.NewReader(destBuf1)
	for {
		hdr1, err := tr1.Next()
		if err == io.EOF {
			fmt.Printf("tr1 eof\n")
			// end of tar archive
			break
		}

		if err == nil {
			fmt.Printf("tr1: %v\n", hdr1.Name)
		} else {
			fmt.Printf("tr1 err: %v\n", err)
			break
		}
	}

	tr2 := tar.NewReader(destBuf2)
	for {
		hdr2, err := tr2.Next()
		if err == io.EOF {
			fmt.Printf("tr2 eof\n")
			// end of tar archive
			break
		}

		if err == nil {
			fmt.Printf("tr2: %v\n", hdr2.Name)
		} else {
			fmt.Printf("tr2 err: %v\n", err)
			break
		}
	}

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
