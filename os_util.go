package elasticthought

import (
	"bufio"
	"io"
	"os"
)

func mkdir(directory string) error {
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
