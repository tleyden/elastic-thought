package elasticthought

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/couchbaselabs/logg"
	"github.com/nu7hatch/gouuid"
)

func Mkdir(directory string) error {
	if err := os.MkdirAll(directory, 0777); err != nil {
		return err
	}
	return nil
}

func streamToFile(r io.Reader, path string) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	defer w.Flush()
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return nil

}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
//
// Source: http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
func CopyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func NewUuid() string {
	u4, err := uuid.NewV4()
	if err != nil {
		logg.LogPanic("Error generating uuid", err)
	}
	return fmt.Sprintf("%s", u4)
}

func saveCmdOutputToFiles(cmdStdout, cmdStderr io.ReadCloser, stdOutPath, stdErrPath string) error {

	stdOutDoneChan := make(chan error, 1)
	stdErrDoneChan := make(chan error, 1)

	// also, Tee everything to this processes' stdout/stderr
	cmdStderrTee := io.TeeReader(cmdStderr, os.Stderr)
	cmdStdoutTee := io.TeeReader(cmdStdout, os.Stdout)

	// spawn goroutines to read from stdout/stderr
	go func() {
		if err := streamToFile(cmdStdoutTee, stdOutPath); err != nil {
			stdOutDoneChan <- err
		} else {
			stdOutDoneChan <- nil
		}

	}()

	go func() {
		if err := streamToFile(cmdStderrTee, stdErrPath); err != nil {
			stdErrDoneChan <- err
		} else {
			stdErrDoneChan <- nil
		}

	}()

	// wait for goroutines
	stdOutResult := <-stdOutDoneChan
	stdErrResult := <-stdErrDoneChan

	// check for errors
	results := []error{stdOutResult, stdErrResult}
	for _, result := range results {
		if result != nil {
			return fmt.Errorf("Saving cmd output failed: %v", result)
		}
	}

	return nil
}

func runCmdTeeStdio(cmd *exec.Cmd, stdOutPath, stdErrPath string) error {

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Error running caffe: StdoutPipe(). Err: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Error running caffe: StderrPipe(). Err: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Error running caffe: cmd.Start(). Err: %v", err)
	}

	// read from stdout, stderr and write to temp files
	if err := saveCmdOutputToFiles(stdout, stderr, stdOutPath, stdErrPath); err != nil {
		return fmt.Errorf("Error running command: saveCmdOutput. Err: %v", err)
	}

	// wait for the command to complete
	runCommandErr := cmd.Wait()

	return runCommandErr

}

func containsString(values []string, testValue string) bool {
	for _, v := range values {
		if v == testValue {
			return true
		}
	}
	return false

}

func getUrlContent(url string) ([]byte, error) {

	// open stream to source url
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error doing GET on: %v.  %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%v response to GET on: %v", resp.StatusCode, url)
	}

	sourceBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading body from: %v.  %v", url, err)
	}

	return sourceBytes, nil

}
