package elasticthought

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/couchbaselabs/cbfs/client"
	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// A solver can generate trained models, which ban be used to make predictions
type Solver struct {
	ElasticThoughtDoc
	DatasetId        string `json:"dataset-id"`
	SpecificationUrl string `json:"specification-url" binding:"required"`
}

// Create a new solver.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewSolver() *Solver {
	return &Solver{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_SOLVER},
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

// download contents of solver-spec-url into cbfs://<solver-id>/spec.prototxt
// and update solver object's solver-spec-url with cbfs url
func (s Solver) SaveSpec(db couch.Database, cbfs *cbfsclient.Client) (*Solver, error) {

	// open stream to source url
	url := s.SpecificationUrl
	resp, err := http.Get(url)
	if err != nil {
		errMsg := fmt.Errorf("Error doing GET on: %v.  %v", url, err)
		return nil, errMsg
	}
	defer resp.Body.Close()

	// save to cbfs
	options := cbfsclient.PutOptions{
		ContentType: "text/plain",
	}
	destPath := fmt.Sprintf("%v/spec.prototxt", s.Id)
	if err := cbfs.Put("", destPath, resp.Body, options); err != nil {
		errMsg := fmt.Errorf("Error writing %v to cbfs: %v", destPath, err)
		return nil, errMsg
	}
	logg.LogTo("REST", "Wrote %v to cbfs", destPath)

	// update solver with cbfs url
	cbfsUrl := fmt.Sprintf("%v%v", CBFS_URI_PREFIX, destPath)
	s.SpecificationUrl = cbfsUrl

	// save
	solver, err := s.Save(db)
	if err != nil {
		return nil, err
	}

	return solver, nil
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
func (s Solver) SaveSpecification(config Configuration, destDirectory string) error {

	// strip leading cbfs://
	specUrlPath, err := s.SpecificationUrlPath()
	if err != nil {
		return err
	}

	// get filename, eg, if path is foo/spec.txt, get spec.txt
	_, specUrlFilename := filepath.Split(specUrlPath)

	// use cbfs client to open stream

	cbfs, err := cbfsclient.New(config.CbfsUrl)
	if err != nil {
		return err
	}

	// get from cbfs
	logg.LogTo("TRAINING_JOB", "Cbfs get %v", specUrlPath)
	reader, err := cbfs.Get(specUrlPath)
	if err != nil {
		return err
	}

	// write stream to file in work directory
	destPath := filepath.Join(destDirectory, specUrlFilename)
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

// If spefication url is "cbfs://foo/bar.txt", return "/foo/bar.txt"
func (s Solver) SpecificationUrlPath() (string, error) {

	specUrl := s.SpecificationUrl
	if !strings.HasPrefix(specUrl, CBFS_URI_PREFIX) {
		return "", fmt.Errorf("Expected %v to start with %v", specUrl, CBFS_URI_PREFIX)
	}

	return strings.Replace(specUrl, CBFS_URI_PREFIX, "", 1), nil

}
