package elasticthought

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/cbfs/client"
)

func saveFileToCbfs(sourcePath, destPath, contentType string, cbfs *cbfsclient.Client) error {

	options := cbfsclient.PutOptions{
		ContentType: contentType,
	}

	f, err := os.Open(sourcePath)
	if err != nil {
		return err
	}

	r := bufio.NewReader(f)

	if err := cbfs.Put("", destPath, r, options); err != nil {
		return fmt.Errorf("Error writing %v to cbfs: %v", destPath, err)
	}

	logg.LogTo("CBFS", "Wrote %v to cbfs: %v", sourcePath, destPath)

	return nil

}

// Save the contents of sourceUrl to cbfs at destPath
func saveUrlToCbfs(sourceUrl, destPath string, cbfs *cbfsclient.Client) error {

	// open stream to source url
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return fmt.Errorf("Error doing GET on: %v.  %v", sourceUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%v response to GET on: %v", resp.StatusCode, sourceUrl)
	}

	options := cbfsclient.PutOptions{
		ContentType: resp.Header.Get("Content-Type"),
	}

	if err := cbfs.Put("", destPath, resp.Body, options); err != nil {
		return fmt.Errorf("Error writing %v to cbfs: %v", destPath, err)
	}

	logg.LogTo("CBFS", "Wrote %v to cbfs: %v", sourceUrl, destPath)

	return nil

}

// Get the content from cbfs from given sourcepath
func getContentFromCbfs(cbfs *cbfsclient.Client, sourcePath string) ([]byte, error) {

	// read contents from cbfs
	reader, err := cbfs.Get(sourcePath)
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
