package elasticthought

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/couchbaselabs/logg"
	"github.com/golang/protobuf/proto"

	"github.com/tleyden/elastic-thought/caffe"
	"github.com/tleyden/go-couch"
)

// A solver can generate trained models, which ban be used to make predictions
type Solver struct {
	ElasticThoughtDoc
	DatasetId           string `json:"dataset-id"`
	SpecificationUrl    string `json:"specification-url" binding:"required"`
	SpecificationNetUrl string `json:"specification-net-url" binding:"required"`

	// had to make exported, due to https://github.com/gin-gonic/gin/pull/123
	// waiting for this to get merged into master branch, since go get
	// pulls from master branch.
	Configuration Configuration

	// distinguish between IMAGE_DATA, LEVELDB, LMDB, etc.
	// this assumes that test and training input layers are
	// of the same layer type.
	LayerType LayerType
}

type LayerType int32

const (
	IMAGE_DATA = LayerType(caffe.V1LayerParameter_IMAGE_DATA)
	DATA       = LayerType(caffe.V1LayerParameter_DATA)
)

// Create a new solver.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewSolver(config Configuration) *Solver {
	return &Solver{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_SOLVER},
		Configuration:     config,
	}
}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (s Solver) Insert(db couch.Database) (*Solver, error) {

	id, _, err := db.Insert(s)
	if err != nil {
		err := fmt.Errorf("Error inserting solver: %v.  Err: %v", s, err)
		return nil, err
	}

	// load dataset object from db (so we have id/rev fields)
	solver := &Solver{}
	err = db.Retrieve(id, solver)
	if err != nil {
		err := fmt.Errorf("Error fetching solver: %v.  Err: %v", id, err)
		return nil, err
	}

	return solver, nil

}

// read solver prototxt from cbfs
func (s Solver) getSolverPrototxtContent() ([]byte, error) {

	// get the relative url path in cbfs (chop off leading cbfs://)
	sourcePath, err := s.SpecificationUrlPath()
	if err != nil {
		return nil, fmt.Errorf("Error getting cbfs path of solver prototxt. Err: %v", err)
	}

	// create a new blob store client
	blobStore, err := s.Configuration.NewBlobStoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error creating blob store client: %v", err)
	}

	return getContentFromBlobStore(blobStore, sourcePath)

}

// read solver prototxt from cbfs
func (s Solver) getSolverNetPrototxtContent() ([]byte, error) {

	// get the relative url path in cbfs (chop off leading cbfs://)
	sourcePath, err := s.SpecificationNetUrlPath()
	if err != nil {
		return nil, fmt.Errorf("Error getting cbfs path of solver prototxt. Err: %v", err)
	}

	// create a new blob store client
	blobStore, err := s.Configuration.NewBlobStoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error creating blob store client: %v", err)
	}

	return getContentFromBlobStore(blobStore, sourcePath)

}

func (s Solver) getSolverParameter() (*caffe.SolverParameter, error) {

	specContents, err := s.getSolverPrototxtContent()
	if err != nil {
		return nil, fmt.Errorf("Error getting solver prototxt content.  Err: %v", err)
	}

	// read into object with protobuf (must have already generated go protobuf code)
	solverParam := &caffe.SolverParameter{}

	if err := proto.UnmarshalText(string(specContents), solverParam); err != nil {
		return nil, err
	}

	return solverParam, nil

}

func (s Solver) getSolverNetParameter() (*caffe.NetParameter, error) {

	specContents, err := s.getSolverNetPrototxtContent()
	if err != nil {
		return nil, fmt.Errorf("Error getting solver prototxt content.  Err: %v", err)
	}

	// read into object with protobuf (must have already generated go protobuf code)
	netParam := &caffe.NetParameter{}

	if err := proto.UnmarshalText(string(specContents), netParam); err != nil {
		return nil, err
	}

	return netParam, nil

}

// download contents of solver-spec-url and make the following modifications:
// - Replace net with "solver-net.prototxt"
// - Replace snapshot_prefix with "snapshot"
func (s Solver) getModifiedSolverSpec() ([]byte, error) {

	// read in spec from url -> []byte
	content, err := getUrlContent(s.SpecificationUrl)
	if err != nil {
		return nil, fmt.Errorf("Error getting data: %v.  %v", s.SpecificationUrl, err)
	}

	// pass in []byte to modifier and get modified []byte
	modified, err := modifySolverSpec(content)
	if err != nil {
		return nil, fmt.Errorf("Error modifying: %v.  %v", string(content), err)
	}

	return modified, nil

}

func modifySolverSpec(source []byte) ([]byte, error) {

	// read into object with protobuf (must have already generated go protobuf code)
	solverParam := &caffe.SolverParameter{}

	if err := proto.UnmarshalText(string(source), solverParam); err != nil {
		return nil, err
	}

	// modify object fields
	solverParam.Net = proto.String("solver-net.prototxt")
	solverParam.SnapshotPrefix = proto.String("snapshot")

	buf := new(bytes.Buffer)
	if err := proto.MarshalText(buf, solverParam); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

// download contents of solver-spec-net-url and make the following modifications:
// - Replace layers / image_data_param / source with "train" and "test"
func (s Solver) getModifiedSolverNetSpec() ([]byte, error) {

	// read in spec from url -> []byte
	content, err := getUrlContent(s.SpecificationNetUrl)
	if err != nil {
		return nil, fmt.Errorf("Error getting data: %v.  %v", s.SpecificationNetUrl, err)
	}

	// pass in []byte to modifier and get modified []byte
	modified, err := modifySolverNetSpec(content)
	if err != nil {
		return nil, fmt.Errorf("Error modifying: %v.  %v", string(content), err)
	}

	return modified, nil

}

func modifySolverNetSpec(sourceBytes []byte) ([]byte, error) {

	// read into object with protobuf (must have already generated go protobuf code)
	netParam := &caffe.NetParameter{}

	if err := proto.UnmarshalText(string(sourceBytes), netParam); err != nil {
		return nil, err
	}

	// modify object fields
	for _, layerParam := range netParam.Layers {

		switch *layerParam.Type {
		case caffe.V1LayerParameter_IMAGE_DATA:

			if isTrainingPhase(layerParam) {
				layerParam.ImageDataParam.Source = proto.String(TRAINING_INDEX)
			}
			if isTestingPhase(layerParam) {
				layerParam.ImageDataParam.Source = proto.String(TESTING_INDEX)
			}

		case caffe.V1LayerParameter_DATA:

			if isTrainingPhase(layerParam) {
				layerParam.DataParam.Source = proto.String(TRAINING_DIR)
			}
			if isTestingPhase(layerParam) {
				layerParam.DataParam.Source = proto.String(TESTING_DIR)
			}

		}

	}

	buf := new(bytes.Buffer)
	if err := proto.MarshalText(buf, netParam); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

func isTrainingPhase(layer *caffe.V1LayerParameter) bool {
	for _, includedPhase := range layer.Include {
		if *includedPhase.Phase == caffe.Phase_TRAIN {
			return true
		}
	}
	return false
}

func isTestingPhase(layer *caffe.V1LayerParameter) bool {
	for _, includedPhase := range layer.Include {
		if *includedPhase.Phase == caffe.Phase_TEST {
			return true
		}
	}
	return false
}

// download contents of solver-spec-url into cbfs://<solver-id>/spec.prototxt
// and update solver object's solver-spec-url with cbfs url
func (s Solver) DownloadSpecToBlobStore(db couch.Database, blobStore BlobStore) (*Solver, error) {

	// rewrite the solver specification
	solverSpecBytes, err := s.getModifiedSolverSpec()
	if err != nil {
		return nil, err
	}

	// save rewritten solver to blobStore
	destPath := fmt.Sprintf("%v/solver.prototxt", s.Id)
	reader := bytes.NewReader(solverSpecBytes)
	if err := s.saveToBlobStore(blobStore, destPath, reader); err != nil {
		return nil, err
	}

	// update solver with blobStore url
	s.SpecificationUrl = fmt.Sprintf("%v%v", CBFS_URI_PREFIX, destPath)

	// rewrite the solver net specification
	solverSpecNetBytes, err := s.getModifiedSolverNetSpec()
	if err != nil {
		return nil, err
	}

	// save rewritten solver to blobStore
	destPath = fmt.Sprintf("%v/solver-net.prototxt", s.Id)
	reader = bytes.NewReader(solverSpecNetBytes)
	if err := s.saveToBlobStore(blobStore, destPath, reader); err != nil {
		return nil, err
	}

	// update solver-net with blobStore url
	s.SpecificationNetUrl = fmt.Sprintf("%v%v", CBFS_URI_PREFIX, destPath)

	// find out whether this is IMAGE_DATA or DATA (leveldb, etc)
	netParam, err := s.getSolverNetParameter()
	if err != nil {
		return nil, err
	}
	s.LayerType = extractTrainingLayerType(netParam)

	// save
	solver, err := s.Save(db)
	if err != nil {
		return nil, err
	}

	return solver, nil
}

func extractTrainingLayerType(netParam *caffe.NetParameter) LayerType {

	// return the layer type of the first layer we see in the net
	// (must be IMAGE_DATA or DATA)
	for _, layerParam := range netParam.Layers {
		return LayerType(*layerParam.Type)
	}

	panic("could not extract training layer type")

}

func (s Solver) saveToBlobStore(blobStore BlobStore, destPath string, reader io.Reader) error {

	// save to blobStore
	options := BlobPutOptions{}
	options.ContentType = "text/plain"

	if err := blobStore.Put("", destPath, reader, options); err != nil {
		return fmt.Errorf("Error writing %v to blobStore: %v", destPath, err)
	}
	logg.LogTo("REST", "Wrote %v to blobStore", destPath)
	return nil

}

// Saves the solver to the db, returns latest rev
func (s Solver) Save(db couch.Database) (*Solver, error) {

	// TODO: retry if 409 error
	_, err := db.Edit(s)
	if err != nil {
		return nil, err
	}

	// load latest version of dataset to return
	solver := &Solver{}
	err = db.Retrieve(s.Id, solver)
	if err != nil {
		return nil, err
	}

	return solver, nil

}

// Save the content in the SpecificationUrl to the given directory.
// As the filename, use the last part of the url path from the SpecificationUrl
func (s Solver) writeSpecToFile(config Configuration, destDirectory string) error {

	// specification
	specUrlPath, err := s.SpecificationUrlPath()
	if err != nil {
		return err
	}
	if err := s.writeBlobStoreFile(config, destDirectory, specUrlPath); err != nil {
		return err
	}

	// specification-net
	specNetUrlPath, err := s.SpecificationNetUrlPath()
	if err != nil {
		return err
	}
	if err := s.writeBlobStoreFile(config, destDirectory, specNetUrlPath); err != nil {
		return err
	}

	return nil
}

// Get a file from cbfs and write it locally
func (s Solver) writeBlobStoreFile(config Configuration, destDirectory, sourceUrl string) error {

	// get filename, eg, if path is foo/spec.txt, get spec.txt
	_, sourceFilename := filepath.Split(sourceUrl)

	// use cbfs client to open stream

	cbfs, err := NewBlobStore(config.CbfsUrl)
	if err != nil {
		return err
	}

	// get from cbfs
	logg.LogTo("TRAINING_JOB", "Cbfs get %v", sourceUrl)
	reader, err := cbfs.Get(sourceUrl)
	if err != nil {
		return err
	}

	// write stream to file in work directory
	destPath := filepath.Join(destDirectory, sourceFilename)
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	defer w.Flush()
	_, err = io.Copy(w, reader)
	if err != nil {
		return err
	}

	logg.LogTo("TRAINING_JOB", "Wrote to %v", destPath)

	return nil

}

// Download and untar the training and test .tar.gz files associated w/ solver,
// as well as index files.
//
// Returns the label index (each label indexed by its numeric label id), and
// an error or nil
func (s Solver) SaveTrainTestData(config Configuration, destDirectory string) (trainingLabelIndex []string, err error) {

	// find cbfs paths to artificacts
	dataset := NewDataset(config)
	dataset.Id = s.DatasetId
	trainingArtifact := dataset.TrainingArtifactPath()
	testArtifact := dataset.TestingArtifactPath()
	trainingLabelIndex = []string{}

	// create blob store client
	blobStore, err := NewBlobStore(config.CbfsUrl)
	if err != nil {
		return nil, err
	}

	// for both the training and testing datafile aka "artifact" (a gzip file stored in cbfs)
	// do the following:
	// - extract it to appropriate subdirectory in destDirectory
	// - find the toc (list of all files in the datafile)
	// - write the toc file
	// - from the toc file of the training set, extract the labelindex
	//
	artificactPaths := []string{trainingArtifact, testArtifact}
	for _, artificactPath := range artificactPaths {

		// open stream to artifact in cbfs
		logg.LogTo("TRAINING_JOB", "Cbfs get %v", artificactPath)
		reader, err := blobStore.Get(artificactPath)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		// Since I'm seeing errors when calling untarGzWithToc:
		//     Err: gzip: invalid header
		// Use a TeeReader to save the raw contents to a file
		_, filename := path.Split(artificactPath)
		destFile := path.Join(destDirectory, filename)
		logg.LogTo("TRAINING_JOB", "Using TeeReader to save copy to %v", destFile)
		f, err := os.Create(destFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		teeReader := io.TeeReader(reader, f)

		subdirectory := ""
		destTocFile := ""
		if artificactPath == trainingArtifact {
			subdirectory = TRAINING_DIR
			destTocFile = path.Join(destDirectory, TRAINING_INDEX)
		} else {
			subdirectory = TESTING_DIR
			destTocFile = path.Join(destDirectory, TESTING_INDEX)
		}
		destDirectoryToUse := path.Join(destDirectory, subdirectory)

		toc, err := untarGzWithToc(teeReader, destDirectoryToUse)
		if err != nil {
			return nil, err
		}
		log.Printf("toc: %v", toc)

		switch s.LayerType {
		case IMAGE_DATA:
			tocWithLabels, labelIndex := addLabelsToToc(toc)
			tocWithSubdir := addParentDirToToc(tocWithLabels, subdirectory)

			if artificactPath == trainingArtifact {
				trainingLabelIndex = labelIndex
			}

			writeTocToFile(tocWithSubdir, destTocFile)

		case DATA:

			// it seems like there is actually no need to do anything
			// when its leveldb, because we don't need either:
			// - the toc
			// - the label mapping

			/*
				// TODO: this doesn't work with leveldb!  see
				// https://github.com/tleyden/elastic-thought/issues/4
				// for leveldb, it needs to generate the labelindex in a
				// different manner.  it needs to just call something to
				// extract the labels from leveldb directly.  (in a separate db)
				log.Printf("Process leveldb: %v", destDirectoryToUse)

				options := &opt.Options{
					ErrorIfMissing: true,
				}
				db, err := leveldb.OpenFile(destDirectoryToUse, options)
				if err != nil {
					return nil, err
				}
				defer db.Close()
				log.Printf("Opened leveldb: %v", db)
				iter := db.NewIterator(nil, nil)
				defer iter.Release()
				for iter.Next() {
					// Remember that the contents of the returned slice should not be modified, and
					// only valid until the next call to Next.
					key := iter.Key()
					// value := iter.Value()
					log.Printf("key: %v", string(key))

					// value: read in protobuf binary into Datum
					datum := &caffe.Datum{}
					err = proto.Unmarshal(iter.Value(), datum)
					if err != nil {
						return nil, err
					}
					log.Printf("datum.label: %v", *datum.Label)

				}
			*/

		}

	}

	// TODO: make sure trainingLabelIndex == testLabelIndex

	return trainingLabelIndex, nil

}

func writeTocToFile(toc []string, destFile string) error {
	f, err := os.Create(destFile)
	if err != nil {
		logg.LogTo("SOLVER", "calling os.Create failed on %v", destFile)
		return err
	}
	w := bufio.NewWriter(f)
	defer w.Flush()
	for _, tocEntry := range toc {
		tocEntryNewline := fmt.Sprintf("%v\n", tocEntry)
		if _, err := w.WriteString(tocEntryNewline); err != nil {
			return err
		}
	}

	return nil
}

/*
Given a toc:

    Q/Verdana-5-0.png 27
    R/Arial-5-0.png 28

And a parent dir, eg, "training-data", generate a new TOC:

    training-data/Q/Verdana-5-0.png 27
    training-data/R/Arial-5-0.png 28

*/
func addParentDirToToc(tableOfContents []string, dir string) []string {

	tocWithDir := []string{}
	for _, tocEntry := range tableOfContents {
		components := strings.Split(tocEntry, " ")
		file := components[0]
		label := components[1]
		file = path.Join(dir, file)
		line := fmt.Sprintf("%v %v", file, label)
		tocWithDir = append(tocWithDir, line)
	}

	return tocWithDir

}

/*
Given a toc:

    Q/Verdana-5-0.png
    R/Arial-5-0.png

Add a numeric label to each line, eg:

    Q/Verdana-5-0.png 27
    R/Arial-5-0.png 28

Where the label starts at 0 and is incremented for
each new directory found.

Return the new toc with numeric labels, followed by the label index.

*/
func addLabelsToToc(tableOfContents []string) (tocWithLabels []string, labels []string) {

	currentDirectory := ""
	labelIndex := 0
	tocWithLabels = []string{}
	labels = []string{}

	for _, tocEntry := range tableOfContents {

		dir := path.Dir(tocEntry)

		if currentDirectory == "" {
			// we're on the first directory
			currentDirectory = dir
		} else {
			// we're not on the first directory, but
			// are we on a new directory?
			if dir == currentDirectory {
				// nope, use curentLabelIndex
			} else {
				// yes, so increment label index
				labelIndex += 1
			}
			currentDirectory = dir
		}

		if !containsString(labels, dir) {
			labels = append(labels, dir)
		}

		tocEntryWithLabel := fmt.Sprintf("%v %v", tocEntry, labelIndex)
		tocWithLabels = append(tocWithLabels, tocEntryWithLabel)

	}

	return tocWithLabels, labels

}

// If spefication url is "cbfs://foo/bar.txt", return "/foo/bar.txt"
func (s Solver) SpecificationUrlPath() (string, error) {

	specUrl := s.SpecificationUrl
	if !strings.HasPrefix(specUrl, CBFS_URI_PREFIX) {
		return "", fmt.Errorf("Expected %v to start with %v", specUrl, CBFS_URI_PREFIX)
	}

	return strings.Replace(specUrl, CBFS_URI_PREFIX, "", 1), nil

}

func (s Solver) SpecificationNetUrlPath() (string, error) {

	specUrl := s.SpecificationNetUrl
	if !strings.HasPrefix(specUrl, CBFS_URI_PREFIX) {
		return "", fmt.Errorf("Expected %v to start with %v", specUrl, CBFS_URI_PREFIX)
	}

	return strings.Replace(specUrl, CBFS_URI_PREFIX, "", 1), nil

}
