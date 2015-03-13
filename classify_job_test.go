package elasticthought

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
	"github.com/tleyden/fakehttp"
)

func TestInvokeCaffe(t *testing.T) {

	tempDir := os.TempDir()

	// cd to temp dir
	err := os.Chdir(tempDir)
	assert.True(t, err == nil)

	// write a fake classifier.py
	python := `
import json
result = {}
result["image4434"] = "5"
result["image7434"] = "A"
f = open('result.json', 'w')
json.dump(result, f)

`

	err = ioutil.WriteFile(filepath.Join(tempDir, "classifier.py"), []byte(python), 0644)
	assert.True(t, err == nil)

	// create new classify job with custom work dir as temp dir
	configuration := NewDefaultConfiguration()
	configuration.WorkDirectory = tempDir
	logg.LogTo("TEST", "temp dir: %v", tempDir)
	classifyJob := NewClassifyJob(*configuration)

	classifier := NewClassifier(*configuration)
	classifier.Scale = "255"
	classifier.ImageHeight = "28"
	classifier.ImageWidth = "28"
	classifier.Color = false
	classifier.Gpu = false

	// call invokeCaffe
	saveStdoutCbfs := false
	results, err := classifyJob.invokeCaffe(saveStdoutCbfs, *classifier)
	logg.LogTo("TEST", "classify results: %v.  err: %v", results, err)
	assert.True(t, err == nil)

	// assert that json has what was expected
	assert.Equals(t, results["image4434"], "5")
	assert.Equals(t, results["image7434"], "A")

}

func TestCopyPythonClassifier(t *testing.T) {

	tempDir := os.TempDir()
	configuration := NewDefaultConfiguration()
	classifyJob := NewClassifyJob(*configuration)
	classifyJob.copyPythonClassifier(tempDir)

	resultFile := path.Join(tempDir, "classifier.py")

	err := validatePathExists(resultFile)
	assert.True(t, err == nil)

}

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

func TestTranslateLabels(t *testing.T) {

	results := map[string]string{
		"foo": "1",
		"bar": "5",
	}

	labels := []string{"a", "b", "c", "d", "e", "f", "g"}

	resultsTranslated, err := translateLabels(results, labels)
	assert.True(t, err == nil)
	assert.Equals(t, resultsTranslated["foo"], "b")
	assert.Equals(t, resultsTranslated["bar"], "f")

}

func TestTranslateLabelsError(t *testing.T) {

	results := map[string]string{
		"foo": "1",
		"bar": "5",
	}

	labels := []string{}

	_, err := translateLabels(results, labels)
	assert.True(t, err != nil)

}

func validatePathExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %v", path)
	}
	return nil
}
