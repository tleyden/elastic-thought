package elasticthought

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
	"github.com/golang/protobuf/proto"
	"github.com/tleyden/elastic-thought/caffe"
)

func TestAddLabelsToToc(t *testing.T) {

	var toc = []string{"foo/1.txt", "bar/1.txt", "bar/2.txt"}
	tocWithLabels, labels := addLabelsToToc(toc)
	for _, entry := range tocWithLabels {
		logg.LogTo("TEST", "entry: %v", entry)
	}
	assert.Equals(t, len(tocWithLabels), len(toc))
	assert.True(t, strings.HasSuffix(tocWithLabels[0], "0"))
	assert.True(t, strings.HasSuffix(tocWithLabels[1], "1"))
	assert.True(t, strings.HasSuffix(tocWithLabels[2], "1"))

	assert.Equals(t, len(labels), 2)
	assert.Equals(t, labels[0], "foo")
	assert.Equals(t, labels[1], "bar")

}

func TestAddDirToToc(t *testing.T) {

	var toc = []string{"foo/1.txt 0", "bar/1.txt 1", "bar/2.txt 1"}
	dir := TRAINING_DIR
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

	modifiedBytes, err := modifySolverSpec([]byte(protoText))
	if err != nil {
		logg.LogError(err)
	}
	assert.True(t, err == nil)
	assert.True(t, len(modifiedBytes) != 0)

	logg.LogTo("TEST", "modified prototxt: %v", string(modifiedBytes))

	// instantiate proto object based on modified bytes
	solverParam := &caffe.SolverParameter{}
	err = proto.UnmarshalText(string(modifiedBytes), solverParam)
	assert.True(t, err == nil)
	assert.True(t, solverParam.Net != nil)
	assert.Equals(t, *(solverParam.Net), "solver-net.prototxt")
	assert.Equals(t, *(solverParam.SnapshotPrefix), "snapshot")

}

func TestGetModifiedSolverNetSpec(t *testing.T) {

	protoText := `
	    name: "alpha"
	    layers {
	      name: "alpha"
	      type: IMAGE_DATA
	      top: "data"
	      top: "label"
	      image_data_param {
		source: "will_be_replaced"
		batch_size: 5
	      }
	      include: { phase: TRAIN }
	    }
	    layers {

	      name: "alpha"
	      type: IMAGE_DATA
	      top: "data"
	      top: "label"
	      image_data_param {
		source: "will_be_replaced"
		batch_size: 10
	      }
	      include: { phase: TEST }
	    }

	    layers {
	      name: "conv1"
	      type: CONVOLUTION
	      bottom: "data"
	      top: "conv1"
	      blobs_lr: 1
	      blobs_lr: 2
	      convolution_param {
		num_output: 20
		kernel_size: 5
		stride: 1
		weight_filler {
		  type: "xavier"
		}
		bias_filler {
		  type: "constant"
		}
	      }
	    }`

	modifiedBytes, err := modifySolverNetSpec([]byte(protoText))
	if err != nil {
		logg.LogError(err)
	}
	assert.True(t, err == nil)
	assert.True(t, len(modifiedBytes) != 0)

	// instantiate proto object based on modified bytes
	netParam := &caffe.NetParameter{}
	err = proto.UnmarshalText(string(modifiedBytes), netParam)
	assert.True(t, err == nil)

	for _, layerParam := range netParam.Layers {

		if *layerParam.Type != caffe.V1LayerParameter_IMAGE_DATA {
			continue
		}

		// is this the training phase or testing phase?
		if isTrainingPhase(layerParam) {
			// get the image data param and make sure the source
			// is "training"
			logg.LogTo("TEST", "is training")
			assert.Equals(t, *layerParam.ImageDataParam.Source, TRAINING_INDEX)
		}
		if isTestingPhase(layerParam) {
			logg.LogTo("TEST", "is testing")
			assert.Equals(t, *layerParam.ImageDataParam.Source, TESTING_INDEX)

		}

	}

}

func TestNetParameter(t *testing.T) {

	// this test does nothing, was just trying to figure out
	// the < vs { issue

	layer := &caffe.V1LayerParameter{}
	layer.Name = proto.String("layer")

	netParam := &caffe.NetParameter{}
	netParam.Name = proto.String("net")

	netParam.Layers = []*caffe.V1LayerParameter{layer}

	marshalled := proto.MarshalTextString(netParam)

	logg.LogTo("TEST", "text marshalled: %v", marshalled)

}

func TestSaveTrainTestDataImageData(t *testing.T) {
	assetName := "data-test/alphabet.tar.gz"
	testSaveTrainTestData(t, assetName, IMAGE_DATA)

}

func testSaveTrainTestData(t *testing.T, assetName string, layerType LayerType) {

	configuration := NewDefaultConfiguration()
	configuration.CbfsUrl = "mock://mock"
	solver := NewSolver(*configuration)
	solver.DatasetId = "123"
	solver.LayerType = layerType

	// get the data corresponding to asset name
	alphabetTarGz, err := Asset(assetName)
	assert.True(t, err == nil)

	DefaultMockBlobStore.QueueGetResponse(
		"123/training.tar.gz",
		bytes.NewBuffer(alphabetTarGz),
	)

	DefaultMockBlobStore.QueueGetResponse(
		"123/testing.tar.gz",
		bytes.NewBuffer(alphabetTarGz),
	)

	destDirectory := "/tmp/TestSaveTrainTestDataImageData"
	Mkdir(destDirectory)
	defer func() {
		os.RemoveAll(destDirectory)
	}()

	labelIndex, err := solver.SaveTrainTestData(*configuration, destDirectory)

	assert.True(t, err == nil)

	switch layerType {
	case IMAGE_DATA:
		log.Printf("LabelIndex: %v", labelIndex)
		assert.Equals(t, len(labelIndex), 26+10)
		assert.Equals(t, labelIndex[0], "0")
		assert.Equals(t, labelIndex[35], "Z")

		destTocFile := path.Join(destDirectory, TRAINING_INDEX)
		log.Printf("destToc: %v", destTocFile)
		destTocLines, err := readFile(destTocFile)
		assert.True(t, err == nil)

		assert.Equals(t, destTocLines[0], "training-data/0/Arial-5-0.png 0\n")
		assert.Equals(t, destTocLines[len(destTocLines)-1], "training-data/Z/Verdana-5-0.png 35\n")

	case DATA:
		// no toc written, empty label index
	}

}

func readFile(filename string) ([]string, error) {
	file, err := os.Open(filename) // For read access.
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(file)
	result := []string{}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		result = append(result, line)
	}
	return result, nil
}
