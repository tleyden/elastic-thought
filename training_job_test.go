package elasticthought

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/tleyden/fakehttp"
)

func TestSaveCmdOutputToFiles(t *testing.T) {

	j := NewTrainingJob(*NewDefaultConfiguration())
	j.Id = "training-job"
	j.createWorkDirectory()

	// build command args
	cmdArgs := []string{}
	cmdPath := "pwd"

	cmd := exec.Command(cmdPath, cmdArgs...)

	stdout, err := cmd.StdoutPipe()
	assert.True(t, err == nil)

	stderr, err := cmd.StderrPipe()
	assert.True(t, err == nil)

	err = cmd.Start()
	assert.True(t, err == nil)

	// read from stdout, stderr and write to temp files
	err = saveCmdOutputToFiles(stdout, stderr, j.getStdOutPath(), j.getStdErrPath())
	assert.True(t, err == nil)

	// wait for the command to complete
	err = cmd.Wait()
	assert.True(t, err == nil)

	// we should have a non-empty stdout/stderr files
	stdoutFile := path.Join(j.getWorkDirectory(), "stdout")
	stderrFile := path.Join(j.getWorkDirectory(), "stderr")
	stdoutBytes, err := ioutil.ReadFile(stdoutFile)
	assert.True(t, err == nil)
	assert.True(t, len(stdoutBytes) > 0)
	stderrBytes, err := ioutil.ReadFile(stderrFile)
	assert.True(t, err == nil)
	assert.True(t, len(stderrBytes) == 0)

}

func TestUpdateProcessingState(t *testing.T) {

	testServer := fakehttp.NewHTTPServerWithPort(NextPort())
	testServer.Start()

	// response when go-couch tries to see that the server is up
	testServer.Response(200, jsonHeaders(), `{"version": "fake"}`)

	// response when go-couch check is db exists
	testServer.Response(200, jsonHeaders(), `{"db_name": "db"}`)

	// first update returns 409
	testServer.Response(409, jsonHeaders(), "")

	// response to GET to refresh
	testServer.Response(200, jsonHeaders(), `{"_id": "training_job", "_rev": "bar", "processing-state": "pending"}`)

	// second update succeeds
	testServer.Response(200, jsonHeaders(), `{"id": "training_job", "rev": "bar"}`)

	configuration := NewDefaultConfiguration()
	configuration.DbUrl = fmt.Sprintf("%v/db", testServer.URL)

	trainingJob := NewTrainingJob(*configuration)
	trainingJob.ElasticThoughtDoc.Id = "training_job"
	trainingJob.ElasticThoughtDoc.Revision = "rev"

	trainingJob.ProcessingState = Pending

	ok, err := trainingJob.UpdateProcessingState(Processing)
	assert.True(t, err == nil)
	assert.True(t, ok)

}

func TestUpdateModelUrl(t *testing.T) {

	docId := "training_job"
	expectedTrainedModelUrl := fmt.Sprintf("%v%v/trained.caffemodel", CBFS_URI_PREFIX, docId)

	testServer := fakehttp.NewHTTPServerWithPort(NextPort())
	testServer.Start()

	// response when go-couch tries to see that the server is up
	testServer.Response(200, jsonHeaders(), `{"version": "fake"}`)

	// response when go-couch check is db exists
	testServer.Response(200, jsonHeaders(), `{"db_name": "db"}`)

	// first update returns 409
	testServer.Response(409, jsonHeaders(), "")

	// response to GET to refresh
	testServer.Response(200, jsonHeaders(), `{"_id": "training_job", "_rev": "bar"}`)

	// second update succeeds
	testServer.Response(200, jsonHeaders(), `{"id": "training_job", "rev": "bar"}`)

	configuration := NewDefaultConfiguration()
	configuration.DbUrl = fmt.Sprintf("%v/db", testServer.URL)

	trainingJob := NewTrainingJob(*configuration)
	trainingJob.ElasticThoughtDoc.Id = docId
	trainingJob.ElasticThoughtDoc.Revision = "rev"

	trainingJob.ProcessingState = Pending

	err := trainingJob.updateCaffeModelUrl()

	assert.True(t, err == nil)
	assert.Equals(t, trainingJob.TrainedModelUrl, expectedTrainedModelUrl)

}

func jsonHeaders() map[string]string {
	return map[string]string{"Content-Type": "application/json"}
}
