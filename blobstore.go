package elasticthought

import (
	"fmt"
	"io"
	"log"
	"net/url"
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

func NewBlobStore(rawurl string) (BlobStore, error) {

	// Three types of blob stores are supported:
	// http://ip:port  (cbfs)
	// mock://mock (mock)
	// file:///path/to/dir (local filesystem)

	url, err := url.Parse(rawurl)
	log.Printf("url: %v, scheme: %v, path: %v, err: %v", url, url.Scheme, url.Path, err)

	switch url.Scheme {
	case "http":
		return NewCbfsBlobStore(rawurl)
	case "mock":
		return NewMockBlobStore(), nil
	case "file":
		return NewFileSystemBlobStore(url.Path)
	default:
		msg := "Unrecognized blob store URI: %v.  If you are trying " +
			"to use cbfs, make sure it ends with port 8484. " +
			"or fix this code"
		return nil, fmt.Errorf(msg)

	}

}
