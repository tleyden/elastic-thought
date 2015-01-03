package elasticthought

import (
	"fmt"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/tleyden/fakehttp"
)

func TestUpdateClassifyJobProcessingState(t *testing.T) {

	testServer := fakehttp.NewHTTPServerWithPort(NextPort())
	testServer.Start()

	// response when go-couch tries to see that the server is up
	testServer.Response(200, jsonHeaders(), `{"version": "fake"}`)

	// response when go-couch check is db exists
	testServer.Response(200, jsonHeaders(), `{"db_name": "db"}`)

	// first update returns 409
	testServer.Response(409, jsonHeaders(), "")

	// response to GET to refresh
	testServer.Response(200, jsonHeaders(), `{"_id": "classify_job", "_rev": "rev2", "processing-state": "pending"}`)

	// second update succeeds
	testServer.Response(200, jsonHeaders(), `{"id": "classify_job", "rev": "rev3"}`)

	configuration := NewDefaultConfiguration()
	configuration.DbUrl = fmt.Sprintf("%v/db", testServer.URL)

	classifyJob := NewClassifyJob(*configuration)
	classifyJob.Id = "classify_job"
	classifyJob.Revision = "rev1"
	classifyJob.ClassifierID = "123"
	classifyJob.ProcessingState = Pending

	updated, err := classifyJob.UpdateProcessingState(Processing)
	assert.True(t, updated)
	assert.True(t, err == nil)
	assert.Equals(t, classifyJob.ProcessingState, Processing)

}
