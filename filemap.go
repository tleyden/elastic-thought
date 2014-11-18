package elasticthought

import "strings"

type filemap map[string][]string

func (f filemap) addFileToDirectory(directory, fileToAdd string) {
	files, ok := f[directory]
	if !ok {
		files = []string{}
		f[directory] = files
	}
	files = append(files, fileToAdd)
	f[directory] = files
}

func (f filemap) hasPath(path string) bool {
	pathComponents := strings.Split(path, "/")
	directory := pathComponents[0]
	filename := pathComponents[1]
	files, ok := f[directory]
	if !ok {
		return false
	}
	for _, file := range files {
		if file == filename {
			return true
		}
	}
	return false
}
