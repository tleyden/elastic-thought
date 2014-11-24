package elasticthought

import (
	"strings"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
)

func TestAddLabelsToToc(t *testing.T) {

	var toc = []string{"foo/1.txt", "bar/1.txt", "bar/2.txt"}
	tocWithLabels := addLabelsToToc(toc)
	for _, entry := range tocWithLabels {
		logg.LogTo("TEST", "entry: %v", entry)
	}
	assert.Equals(t, len(tocWithLabels), len(toc))
	assert.True(t, strings.HasSuffix(tocWithLabels[0], "0"))
	assert.True(t, strings.HasSuffix(tocWithLabels[1], "1"))
	assert.True(t, strings.HasSuffix(tocWithLabels[2], "1"))

}

func TestAddDirToToc(t *testing.T) {

	var toc = []string{"foo/1.txt 0", "bar/1.txt 1", "bar/2.txt 1"}
	dir := "training-data"
	tocWithDirs := addParentDirToToc(toc, dir)
	for _, entry := range tocWithDirs {
		logg.LogTo("TEST", "entry: %v", entry)
	}
	assert.Equals(t, len(tocWithDirs), len(toc))
	for _, tocEntry := range tocWithDirs {
		assert.True(t, strings.HasPrefix(tocEntry, dir))
	}

}
