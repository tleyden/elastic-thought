package elasticthought

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/couchbaselabs/logg"
)

type FileSystemBlobStore struct {

	// the root path where blobs will be stored.  all paths will
	// be calculated relative to this path.
	rootPath string
}

type FileSystemBlobHandle struct{}

func (h FileSystemBlobHandle) Nodes() map[string]time.Time {
	return map[string]time.Time{
		"filesystem": time.Now(),
	}
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

func (f FileSystemBlobStore) Get(path string) (io.ReadCloser, error) {

	// build full path to file
	path = f.absolutePath(path)

	// open the file for reading and return
	return os.Open(path)

}

func (f FileSystemBlobStore) Put(srcname, dest string, r io.Reader, opts BlobPutOptions) error {

	logg.LogTo("ELASTIC_THOUGHT", "FileSystemBlobStore.Put() called with: %v", dest)

	// build full path to file
	path := f.absolutePath(dest)

	// create parent dir path if it doesn't already exist
	logg.LogTo("ELASTIC_THOUGHT", "dir: %v", filepath.Dir(path))
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))

	// open the file for writing
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	// copy from reader -> file
	_, err = io.Copy(file, r)

	return err
}

func (f FileSystemBlobStore) Rm(path string) error {

	// build full path to file
	path = f.absolutePath(path)

	return os.Remove(path)

}

func (f FileSystemBlobStore) OpenFile(path string) (BlobHandle, error) {
	return FileSystemBlobHandle{}, nil
}

func (f FileSystemBlobStore) absolutePath(relativePath string) string {
	return filepath.Join(f.rootPath, relativePath)
}
