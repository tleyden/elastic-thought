package elasticthought

import (
	"strings"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
	"github.com/golang/protobuf/proto"
	"github.com/tleyden/elastic-thought/caffe"
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

func TestGetModifiedSolverSpec(t *testing.T) {

	protoText := `
	    # The train/test net protocol buffer definition
	    net: "this_should_get_replaced"
	    # test_iter specifies how many forward passes the test should carry out.
	    # In the case of MNIST, we have test batch size 100 and 100 test iterations,
	    # covering the full 10,000 testing images.
	    test_iter: 100
	    # Carry out testing every 500 training iterations.
	    test_interval: 500
	    # The base learning rate, momentum and the weight decay of the network.
	    base_lr: 0.01
	    momentum: 0.9
	    weight_decay: 0.0005
	    # The learning rate policy
	    lr_policy: "inv"
	    gamma: 0.0001
	    power: 0.75
	    # Display every 100 iterations
	    display: 100
	    # The maximum number of iterations
	    max_iter: 10000
	    # snapshot intermediate results
	    snapshot: 5000
	    snapshot_prefix: "snapshot"
	    # solver mode: CPU or GPU
	    solver_mode: CPU`

	modifiedBytes, err := getModifiedSolverSpec(protoText)
	if err != nil {
		logg.LogError(err)
	}
	assert.True(t, err == nil)
	assert.True(t, len(modifiedBytes) != 0)

	// instantiate proto object based on modified bytes
	solverParam := &caffe.SolverParameter{}
	err = proto.UnmarshalText(string(modifiedBytes), solverParam)
	assert.True(t, err == nil)
	assert.True(t, solverParam.Net != nil)
	assert.Equals(t, *(solverParam.Net), "solver-net.prototxt")

}
