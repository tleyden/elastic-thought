package elasticthought

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/couchbaselabs/cbfs/client"
)

// BlobStore provides a blob store interface.
type BlobStore interface {
	Get(path string) (io.ReadCloser, error)
	Put(srcname, dest string, r io.Reader, opts BlobPutOptions) error
	Rm(fn string) error
	OpenFile(path string) (BlobHandle, error)
}

// Calling OpenFile with a path on a BlobStore returns a
// BlobHandle which allows for reading and metadata.
type BlobHandle interface {

	// The nodes that contain the file corresponding to this blob
	// and the last time it was "scrubbed" (cbfs terminology)
	Nodes() map[string]time.Time
}

type BlobPutOptions struct {
	cbfsclient.PutOptions
}

func NewBlobStore(uri string) (BlobStore, error) {
	if strings.HasSuffix(uri, "8484") {
		cbfsBlobStore, err := NewCbfsBlobStore(uri)
		if err != nil {
			return nil, err
		}
		return cbfsBlobStore, nil
	} else if strings.Contains(uri, "mock-blob-store") {
		return NewMockBlobStore(), nil
	} else {
		msg := "Unrecognized blob store URI: %v.  If you are trying " +
			"to use cbfs, make sure it ends with port 8484. " +
			"or fix this code"
		return nil, fmt.Errorf(msg)
	}

}
