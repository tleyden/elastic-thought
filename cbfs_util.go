package elasticthought

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/couchbaselabs/logg"
)

func saveFileToBlobStore(sourcePath, destPath, contentType string, blobStore BlobStore) error {

	options := BlobPutOptions{}
	options.ContentType = contentType

	f, err := os.Open(sourcePath)
	if err != nil {
		return err
	}

	r := bufio.NewReader(f)

	if err := blobStore.Put("", destPath, r, options); err != nil {
		return fmt.Errorf("Error writing %v to cbfs: %v", destPath, err)
	}

	logg.LogTo("CBFS", "Wrote %v to cbfs: %v", sourcePath, destPath)

	return nil

}

// Save the contents of sourceUrl to cbfs at destPath
func saveUrlToBlobStore(sourceUrl, destPath string, blobStore BlobStore) error {

	// open stream to source url
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return fmt.Errorf("Error doing GET on: %v.  %v", sourceUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%v response to GET on: %v", resp.StatusCode, sourceUrl)
	}

	options := BlobPutOptions{}
	options.ContentType = resp.Header.Get("Content-Type")

	if err := blobStore.Put("", destPath, resp.Body, options); err != nil {
		return fmt.Errorf("Error writing %v to blobStore: %v", destPath, err)
	}

	logg.LogTo("CBFS", "Wrote %v to blobStore: %v", sourceUrl, destPath)

	return nil

}

// Get the content from cbfs from given sourcepath
func getContentFromBlobStore(blobStore BlobStore, sourcePath string) ([]byte, error) {

	// read contents from blobStore
	reader, err := blobStore.Get(sourcePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return bytes, nil

}

// Download the content at sourcePath (cbfs://foo/bar.txt) to destPath (/path/to/bar.txt)
func downloadFromBlobStore(blobStore BlobStore, sourceUri, destPath string) (err error) {

	if !strings.HasPrefix(sourceUri, CBFS_URI_PREFIX) {
		return fmt.Errorf("Invalid TrainedModelUrl: %v", sourceUri)
	}

	// open a file at destPath
	out, err := os.Create(destPath)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	// chop off cbfs:// to get a source path on cbfs
	sourcePath := strings.Replace(sourceUri, CBFS_URI_PREFIX, "", -1)

	// read contents from blobStore
	reader, err := blobStore.Get(sourcePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// copy blobStore -> dest
	_, err = io.Copy(out, reader)

	return

}
