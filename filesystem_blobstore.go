package elasticthought

import (
	"io"
	"os"
)

type FileSystemBlobStore struct {

	// the root path where blobs will be stored.  all paths will
	// be calculated relative to this path.
	rootPath string
}

func NewFileSystemBlobStore(rootPath string) (*FileSystemBlobStore, error) {

	// create rootPath if it doesn't exist
	err := os.MkdirAll(rootPath, os.FileMode(0755))
	if err != nil {
		return nil, err
	}

	fsBlobStore := &FileSystemBlobStore{
		rootPath: rootPath,
	}

	return fsBlobStore, nil

}

func (f *FileSystemBlobStore) Get(path string) (io.ReadCloser, error) {
	return nil, nil
}

func (f *FileSystemBlobStore) Put(srcname, dest string, r io.Reader, opts BlobPutOptions) error {
	return nil
}

func (f *FileSystemBlobStore) Rm(fn string) error {
	return nil
}

func (f *FileSystemBlobStore) OpenFile(path string) (BlobHandle, error) {
	return nil, nil
}
