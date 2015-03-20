package elasticthought

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// A classify job tries to classify images given by user against
// the given trained model
type ClassifyJob struct {
	ElasticThoughtDoc
	ProcessingState ProcessingState `json:"processing-state"`
	ProcessingLog   string          `json:"processing-log"`
	StdOutUrl       string          `json:"std-out-url"`
	StdErrUrl       string          `json:"std-err-url"`
	ClassifierID    string          `json:"classifier-id"`

	// Key: image url of image in cbfs
	// Value: the classification result for that image
	Results map[string]string `json:"results"`

	// had to make exported, due to https://github.com/gin-gonic/gin/pull/123
	// waiting for this to get merged into master branch, since go get
	// pulls from master branch.
	Configuration Configuration
}

// Create a new classify job.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewClassifyJob(c Configuration) *ClassifyJob {
	return &ClassifyJob{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_CLASSIFY_JOB},
		Configuration:     c,
	}
}

// Run this job
func (c *ClassifyJob) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	logg.LogTo("CLASSIFY_JOB", "Run() called!")

	updatedState, err := c.UpdateProcessingState(Processing)
	if err != nil {
		c.recordProcessingError(err)
		return
	}

	if !updatedState {
		logg.LogTo("CLASSIFY_JOB", "%+v already processed.  Ignoring.", c)
		return
	}

	// TODO: add code to run job

	logg.LogTo("CLASSIFY_JOB", "lazily create dir.  images: %+v", c.Results)

	if err := c.createWorkDirectory(); err != nil {
		c.recordProcessingError(err)
		return
	}

	if err := c.downloadWorkAssets(); err != nil {
		c.recordProcessingError(err)
		return
	}

	classifier, err := c.getClassifier()
	if err != nil {
		c.recordProcessingError(err)
		return
	}

	// invoke caffe
	saveStdoutCbfs := false
	resultsMap, err := c.invokeCaffe(saveStdoutCbfs, *classifier)
	if err != nil {
		c.recordProcessingError(err)
		return
	}

	// get the training job
	trainingJob, err := c.getTrainingJob()
	if err != nil {
		c.recordProcessingError(err)
		return
	}

	// get the solver
	solver, err := c.getSolver()
	if err != nil {
		c.recordProcessingError(err)
		return
	}

	switch solver.LayerType {
	case IMAGE_DATA:
		// modify results to map numeric labels with actual labels
		logg.LogTo("CLASSIFY_JOB", "raw results: %+v.", resultsMap)
		resultsMap, err = translateLabels(resultsMap, trainingJob.Labels)
		if err != nil {
			c.recordProcessingError(err)
			return
		}

	case DATA:
		// no label translation needed
	}

	// update classifyjob with results
	logg.LogTo("CLASSIFY_JOB", "resultsMap: %+v", resultsMap)
	_, err = c.SetResults(resultsMap)
	if err != nil {
		c.recordProcessingError(err)
		return
	}

	// set job state to finished
	_, err = c.UpdateProcessingState(FinishedSuccessfully)
	if err != nil {
		c.recordProcessingError(err)
		return
	}

	logg.LogTo("CLASSIFY_JOB", "Finished")

}

// Invoke caffe to do classification and return a map with:
//    <image_sha1>:<numeric_label_string>
//
// Example:
//    {"b56b61d15ccff4a81a4":"9","daf9e2c49ddbee3d48":"14"}
func (c ClassifyJob) invokeCaffe(saveStdoutCbfs bool, classifier Classifier) (map[string]string, error) {

	// build command args for calling "python classifier.py <args>"
	cmdArgs := []string{
		"classifier.py",
		"--scale",
		classifier.Scale,
		"--image-width",
		classifier.ImageWidth,
		"--image-height",
		classifier.ImageHeight,
	}

	if classifier.Color {
		cmdArgs = append(cmdArgs, "--color")
	}
	if classifier.Gpu {
		cmdArgs = append(cmdArgs, "--gpu")
	}

	python := "python"

	// debugging
	logg.LogTo("CLASSIFY_JOB", "Running %v with args %v", python, cmdArgs)
	logg.LogTo("CLASSIFY_JOB", "Path %v", os.Getenv("PATH"))

	// explicitly check if caffe binary found on the PATH
	lookPathResult, err := exec.LookPath("python")
	if err != nil {
		logg.LogError(fmt.Errorf("python not found on path: %v", err))
	}
	logg.LogTo("CLASSIFY_JOB", "python found on path: %v", lookPathResult)

	// Create command, but don't actually run it yet
	cmd := exec.Command(python, cmdArgs...)

	// set the directory where the command will be run in (important
	// because we depend on relative file paths to work)
	cmd.Dir = c.getWorkDirectory()

	// run the command and save stdio to files and tee to stdio streams
	if err := runCmdTeeStdio(cmd, c.getStdOutPath(), c.getStdErrPath()); err != nil {
		return nil, err
	}

	if saveStdoutCbfs {

		// read from temp files and write to cbfs.
		// initially I tried to write the stdout/stderr streams directly
		// to cbfs, but ran into an error related to the io.Seeker interface.
		if err := c.saveCmdOutputToCbfs(c.getStdOutPath()); err != nil {
			return nil, fmt.Errorf("Could not save output to cbfs. Err: %v", err)
		}

		if err := c.saveCmdOutputToCbfs(c.getStdErrPath()); err != nil {
			return nil, fmt.Errorf("Could not save output to cbfs. Err: %v", err)
		}

	}

	// read output.json file into map
	result := map[string]string{}
	resultFilePath := filepath.Join(c.getWorkDirectory(), "result.json")
	resultFile, err := os.Open(resultFilePath)
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(resultFile)
	if err = jsonParser.Decode(&result); err != nil {
		return nil, err
	}

	return result, nil

}

func (c ClassifyJob) getStdOutPath() string {
	return path.Join(c.getWorkDirectory(), "stdout")
}

func (c ClassifyJob) getStdErrPath() string {
	return path.Join(c.getWorkDirectory(), "stderr")
}

func (c ClassifyJob) getWorkDirectory() string {
	return filepath.Join(c.Configuration.WorkDirectory, c.Id)
}

func (c ClassifyJob) getWorkImagesDirectory() string {
	return filepath.Join(c.getWorkDirectory(), "images")
}

func (c ClassifyJob) saveCmdOutputToCbfs(sourcePath string) error {

	base := path.Base(sourcePath)
	destPath := fmt.Sprintf("%v/%v", c.Id, base)

	cbfsclient, err := c.Configuration.NewBlobStoreClient()
	if err != nil {
		return err
	}

	if err := saveFileToBlobStore(sourcePath, destPath, "text/plain", cbfsclient); err != nil {
		return err
	}

	return nil

}

// lazily create work dir
func (c ClassifyJob) createWorkDirectory() error {
	workDir := c.getWorkDirectory()
	logg.LogTo("CLASSIFY_JOB", "Creating dir: %v", workDir)
	if err := Mkdir(workDir); err != nil {
		return err
	}

	imagesSubdir := c.getWorkImagesDirectory()
	if err := Mkdir(imagesSubdir); err != nil {
		return err
	}

	return nil

}

func (c ClassifyJob) getClassifier() (*Classifier, error) {

	classifier := NewClassifier(c.Configuration)
	err := classifier.Find(c.ClassifierID)
	return classifier, err
}

func (c ClassifyJob) getTrainingJob() (*TrainingJob, error) {

	classifier, err := c.getClassifier()
	if err != nil {
		return nil, err
	}

	return classifier.getTrainingJob()

}

func (c ClassifyJob) getSolver() (*Solver, error) {

	trainingJob, err := c.getTrainingJob()
	if err != nil {
		return nil, err
	}
	return trainingJob.getSolver()

}

func (c ClassifyJob) downloadWorkAssets() error {

	// caffe model
	if err := c.downloadCaffeModel(); err != nil {
		return err
	}

	// prototxt
	if err := c.downloadPrototxt(); err != nil {
		return err
	}

	// images
	if err := c.downloadImagesToClassify(); err != nil {
		return err
	}

	// python classify script
	workDirectory := c.getWorkDirectory()
	if err := c.copyPythonClassifier(workDirectory); err != nil {
		return err
	}

	return nil

}

func (c ClassifyJob) copyPythonClassifier(destDirPath string) error {

	// find out $GOPATH env variable
	gopath := os.Getenv("GOPATH")

	// find path to python classifier file
	elasticThoughtRoot := path.Join(
		gopath,
		"src",
		"github.com",
		"tleyden",
		"elastic-thought",
	)

	pythonScript := path.Join(
		elasticThoughtRoot,
		"scripts",
		"python-classifier",
		"classifier.py",
	)

	destPath := path.Join(destDirPath, "classifier.py")

	if err := CopyFileContents(pythonScript, destPath); err != nil {
		return err
	}

	return nil

}

func (c ClassifyJob) downloadCaffeModel() error {

	// get training job id, instantiate object
	trainingJob, err := c.getTrainingJob()
	if err != nil {
		return err
	}

	// make sure that training job state == Finished
	if trainingJob.GetProcessingState() != FinishedSuccessfully {
		return fmt.Errorf("TrainingJob is not finished yet")
	}

	// make sure the trained model url starts with cbfs
	trainedModelUrl := trainingJob.TrainedModelUrl

	// download from cbfs to local file system
	cbfs, err := c.Configuration.NewBlobStoreClient()
	if err != nil {
		return err
	}
	destPath := path.Join(c.getWorkDirectory(), "caffe.model")
	if err := downloadFromBlobStore(cbfs, trainedModelUrl, destPath); err != nil {
		return err
	}

	return nil

}

func (c ClassifyJob) downloadPrototxt() error {

	classifier, err := c.getClassifier()
	if err != nil {
		return err
	}

	cbfs, err := c.Configuration.NewBlobStoreClient()
	if err != nil {
		return err
	}
	destPath := path.Join(c.getWorkDirectory(), "classifier.prototxt")
	if err := downloadFromBlobStore(cbfs, classifier.SpecificationUrl, destPath); err != nil {
		return err
	}

	return nil
}

func (c ClassifyJob) downloadImagesToClassify() error {

	// for now, just download a single image
	// (if multiple images specified, all bust last will be clobbered)
	// TODO: fix this and download all images

	cbfs, err := c.Configuration.NewBlobStoreClient()
	if err != nil {
		return err
	}

	i := 0
	for imageUrl, _ := range c.Results {

		// url will be cbfs://<classify_job_id>/<imageurl_sha1_hash>
		_, imageSha1Hash := path.Split(imageUrl)
		destPath := path.Join(c.getWorkImagesDirectory(), imageSha1Hash)
		if err := downloadFromBlobStore(cbfs, imageUrl, destPath); err != nil {
			return err
		}
		i += 1
	}

	return nil

}

// Update the processing state to new state.
func (c *ClassifyJob) UpdateProcessingState(newState ProcessingState) (bool, error) {

	updater := func(classifyJob *ClassifyJob) {
		classifyJob.ProcessingState = newState
	}

	doneMetric := func(classifyJob ClassifyJob) bool {
		return classifyJob.ProcessingState == newState
	}

	return c.casUpdate(updater, doneMetric)

}

func (c *ClassifyJob) SetResults(results map[string]string) (bool, error) {

	updater := func(classifyJob *ClassifyJob) {
		classifyJob.Results = results
	}

	doneMetric := func(classifyJob ClassifyJob) bool {
		return reflect.DeepEqual(results, classifyJob.Results)
	}

	return c.casUpdate(updater, doneMetric)

}

func (c *ClassifyJob) UpdateProcessingLog(val string) (bool, error) {

	updater := func(classifyJob *ClassifyJob) {
		classifyJob.ProcessingLog = val
	}

	doneMetric := func(classifyJob ClassifyJob) bool {
		return classifyJob.ProcessingLog == val
	}

	return c.casUpdate(updater, doneMetric)

}

func (c *ClassifyJob) casUpdate(updater func(*ClassifyJob), doneMetric func(ClassifyJob) bool) (bool, error) {

	db := c.Configuration.DbConnection()

	genUpdater := func(classifyJobPtr interface{}) {
		cjp := classifyJobPtr.(*ClassifyJob)
		updater(cjp)
	}

	genDoneMetric := func(classifyJobPtr interface{}) bool {
		cjp := classifyJobPtr.(*ClassifyJob)
		return doneMetric(*cjp)
	}

	refresh := func(classifyJobPtr interface{}) error {
		cjp := classifyJobPtr.(*ClassifyJob)
		return cjp.RefreshFromDB(db)
	}

	return casUpdate(db, c, genUpdater, genDoneMetric, refresh)

}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (c *ClassifyJob) Insert() error {

	db := c.Configuration.DbConnection()

	id, rev, err := db.Insert(c)
	if err != nil {
		err := fmt.Errorf("Error inserting classify job: %v.  Err: %v", c, err)
		return err
	}

	c.Id = id
	c.Revision = rev

	return nil

}

// CodeReview: duplication with RefreshFromDB in many places
func (c *ClassifyJob) RefreshFromDB(db couch.Database) error {
	classifyJob := ClassifyJob{}
	err := db.Retrieve(c.Id, &classifyJob)
	if err != nil {
		return err
	}
	*c = classifyJob
	return nil
}

// Find a classify Job in the db with the given id, or error if not found
// CodeReview: duplication with Find in many places
func (c *ClassifyJob) Find(id string) error {
	db := c.Configuration.DbConnection()
	c.Id = id
	if err := c.RefreshFromDB(db); err != nil {
		return err
	}
	return nil
}

// Codereview: de-dupe
func (c ClassifyJob) recordProcessingError(err error) {
	logg.LogError(err)
	db := c.Configuration.DbConnection()
	if err := c.Failed(db, err); err != nil {
		errMsg := fmt.Errorf("Error setting training job as failed: %v", err)
		logg.LogError(errMsg)
	}
}

func (c ClassifyJob) Failed(db couch.Database, processingErr error) error {

	_, err := c.UpdateProcessingState(Failed)
	if err != nil {
		return err
	}

	logg.LogTo("CLASSIFY_JOB", "updating processing log")

	logValue := fmt.Sprintf("%v", processingErr)
	_, err = c.UpdateProcessingLog(logValue)
	if err != nil {
		return err
	}

	return nil

}

// Given {"b56b61d15ccff4a81a4":"9","daf9e2c49ddbee3d48":"14"} return a map
// with numeric labels translated into actual labels.
// Example: {"b56b61d15ccff4a81a4":"9","daf9e2c49ddbee3d48":"E"}
func translateLabels(results map[string]string, labels []string) (map[string]string, error) {

	transformedResults := map[string]string{}

	for imageSha1, numericLabelString := range results {

		// "14" -> 14
		numericLabel, err := strconv.ParseInt(numericLabelString, 10, 64)
		if err != nil {
			return nil, err
		}

		if int(numericLabel) > (len(labels) - 1) {
			return nil, fmt.Errorf("No label at index: %v", numericLabel)
		}

		// 14 -> "E"
		label := labels[numericLabel]

		// store in result
		transformedResults[imageSha1] = label

	}

	return transformedResults, nil

}
