package elasticthought

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
)

// Opens tar.gz stream
func openTarGzStream(url string) (*tar.Reader, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(gzipReader)

	return tarReader, nil

}

// Given a reader, wrap in a tar.gz reader and write all entries
// to destDirectory.  Also return a table of contents.
func untarGzWithToc(reader io.Reader, destDirectory string) ([]string, error) {
	return nil, nil
}
