package elasticthought

import (
	"bufio"
	"fmt"
	"io"
	"os"

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
