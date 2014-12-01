package elasticthought

import (
	"io/ioutil"
	"os/exec"
	"path"
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestSaveCmdOutputToFiles(t *testing.T) {

	j := NewTrainingJob()
	j.Configuration = *NewDefaultConfiguration()
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
	err = j.saveCmdOutputToFiles(stdout, stderr)
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
