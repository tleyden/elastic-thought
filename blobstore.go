package elasticthought

import (
	"fmt"
	"io"
	"strings"

	"github.com/couchbaselabs/cbfs/client"
)

// BlobStore provides a blob store interface.
type BlobStore interface {
	Get(path string) (io.ReadCloser, error)
	Put(srcname, dest string, r io.Reader, opts cbfsclient.PutOptions) error
	Rm(fn string) error
	OpenFile(path string) (*cbfsclient.FileHandle, error)
}

func NewBlobStore(uri string) (BlobStore, error) {
	if strings.HasSuffix(uri, "8484") {
		cbfsClient, err := cbfsclient.New(uri)
		if err != nil {
			return nil, err
		}
		return cbfsClient, nil
	} else if strings.Contains(uri, "mock-blob-store") {
		return NewMockBlobStore(), nil
	} else {
		msg := "Unrecognized blob store URI: %v.  If you are trying " +
			"to use cbfs, make sure it ends with port 8484. " +
			"or fix this code"
		return nil, fmt.Errorf(msg)
	}

}
