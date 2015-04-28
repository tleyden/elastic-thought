package elasticthought

import (
	"io"

	"github.com/couchbaselabs/cbfs/client"
)

type CbfsBlobStore struct {
	uri string
	*cbfsclient.Client
}

func NewCbfsBlobStore(uri string) (*CbfsBlobStore, error) {
	cbfsBlobStore := &CbfsBlobStore{
		uri: uri,
	}
	cbfsClient, err := cbfsclient.New(uri)
	if err != nil {
		return nil, err
	}
	cbfsBlobStore.Client = cbfsClient
	return cbfsBlobStore, nil
}

func (c *CbfsBlobStore) OpenFile(path string) (BlobHandle, error) {
	return c.Client.OpenFile(path)
}

func (c *CbfsBlobStore) Put(srcname, dest string, r io.Reader, opts BlobPutOptions) error {

	cbfsPutOptions := opts.PutOptions

	return c.Client.Put(srcname, dest, r, cbfsPutOptions)
}

/*
func (c *CbfsBlobStore) OpenFile(path string) (BlobHandle, error) {
	cbfsFileHandle, err := c.Client.OpenFile(path)
	if err != nil {
		return nil, err
	}
	blobHandle, ok := cbfsFileHandle.(BlobHandle)
	if !ok {
		return nil, fmt.Errorf("Couldn't convernt cbfs file handle to blob handle")
	}
	return blobHandle
}
*/
