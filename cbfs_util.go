package elasticthought

import (
	"bufio"
	"fmt"
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
