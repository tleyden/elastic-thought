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
