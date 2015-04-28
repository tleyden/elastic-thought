package elasticthought

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestNewBlobStore(t *testing.T) {

	blobStore, err := NewBlobStore("mock://mock")
	assert.True(t, err == nil)
	_, ok := blobStore.(*MockBlobStore)
	assert.True(t, ok)

	fileBlobStore, err := NewBlobStore("file:///tmp")
	assert.True(t, err == nil)
	_, ok = fileBlobStore.(*FileSystemBlobStore)
	assert.True(t, ok)

}
