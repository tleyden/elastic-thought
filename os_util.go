package elasticthought

import "os"

func mkdir(directory string) error {
	if err := os.MkdirAll(directory, 0777); err != nil {
		return err
	}
	return nil
}
