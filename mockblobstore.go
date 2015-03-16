package elasticthought

import (
	"io"

	"github.com/couchbaselabs/cbfs/client"
)

type MockBlobStore struct {

	// Queued responses for blob Get requests.  The key is a
	// regex that should match the path in the Get request.
	GetResponses map[string]ResponseQueue
}

type ResponseQueue []io.Reader

func NewMockBlobStore() *MockBlobStore {
	mbs := &MockBlobStore{
		GetResponses: map[string]ResponseQueue{},
	}
	return mbs
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

func (m *MockBlobStore) Get(path string) (io.ReadCloser, error) {
	// TODO:
	return nil, nil
}

func (m *MockBlobStore) Put(srcname, dest string, r io.Reader, opts cbfsclient.PutOptions) error {
	return nil
}
