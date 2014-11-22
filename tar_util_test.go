package elasticthought

import (
	"bytes"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
)

func TestUntarGzWithToc(t *testing.T) {

	// Create a test tar archive
	buf := new(bytes.Buffer)

	var files = []tarFile{
		{"foo/1.txt", "."},
		{"foo/2.txt", "."},
		{"bar/1.txt", "."},
		{"bar/2.txt", "."},
		{"bar/3.txt", "."},
		{"bar/4.txt", "."},
		{"bar/5.txt", "."},
	}
	createArchive(buf, files)
	reader := bytes.NewReader(buf.Bytes())

	tempDir := TempDir()
	logg.LogTo("TEST", "tempDir: %v", tempDir)
	toc, err := untarWithToc(reader, tempDir)
	assert.True(t, err == nil)

	logg.LogTo("TEST", "toc: %v, err: %v", toc, err)

	// TODO: add asserations

}
