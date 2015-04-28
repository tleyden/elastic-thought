package elasticthought

import (
	"fmt"
	"io"
	"strings"
)

var (
	DefaultMockBlobStore *MockBlobStore
)

func init() {
	DefaultMockBlobStore = &MockBlobStore{
		GetResponses: map[string]ResponseQueue{},
	}
}

type MockBlobStore struct {

	// Queued responses for blob Get requests.  The key is a
	// regex that should match the path in the Get request.
	GetResponses map[string]ResponseQueue
}

type ResponseQueue []io.Reader

func NewMockBlobStore() *MockBlobStore {
	if DefaultMockBlobStore == nil {
		DefaultMockBlobStore = &MockBlobStore{
			GetResponses: map[string]ResponseQueue{},
		}

	}
	return DefaultMockBlobStore
}

func (m *MockBlobStore) Get(path string) (io.ReadCloser, error) {
	matchingKey, queue := m.responseQueueForPath(path)
	if len(queue) == 0 {
		return nil, fmt.Errorf("No more items in mock blob store for %v", path)
	}
	firstItem := queue[0]
	m.GetResponses[matchingKey] = queue[1:]

	return nopCloser{firstItem}, nil
}

func (m *MockBlobStore) Put(srcname, dest string, r io.Reader, opts BlobPutOptions) error {
	return nil
}

func (m *MockBlobStore) Rm(fn string) error {
	return nil
}

func (m *MockBlobStore) OpenFile(path string) (BlobHandle, error) {
	return nil, nil
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

func (m *MockBlobStore) responseQueueForPath(path string) (string, ResponseQueue) {
	// loop over all keys in GetResponses map until we find a match
	for k, v := range m.GetResponses {
		if path == "*" {
			return k, v
		}
		if strings.Contains(k, path) { // TODO: replace w/ regex
			return k, v
		}
	}
	return "", nil
}

// Queue up a response to a Get request
func (m *MockBlobStore) QueueGetResponse(pathRegex string, response io.Reader) {
	queue, ok := m.GetResponses[pathRegex]
	if !ok {
		queue = ResponseQueue{}
	}
	queue = append(queue, response)
	m.GetResponses[pathRegex] = queue
}
