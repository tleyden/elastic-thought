package elasticthought

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/cbfs/client"
	"github.com/tleyden/go-couch"
)

// A training job represents a "training session" of a solver against training/test data
type TrainingJob struct {
	ElasticThoughtDoc
	ProcessingState ProcessingState `json:"processing-state"`
	ProcessingLog   string          `json:"processing-log"`
	UserID          string          `json:"user-id"`
	SolverId        string          `json:"solver-id" binding:"required"`
	StdOutUrl       string          `json:"std-out-url"`
	StdErrUrl       string          `json:"std-err-url"`
	TrainedModelUrl string          `json:"trained-model-url"`

	// had to make exported, due to https://github.com/gin-gonic/gin/pull/123
	// waiting for this to get merged into master branch, since go get
	// pulls from master branch.
	Configuration Configuration
}

// Create a new training job.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewTrainingJob() *TrainingJob {
	return &TrainingJob{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_TRAINING_JOB},
	}
}

// Run this job
func (j TrainingJob) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	logg.LogTo("TRAINING_JOB", "Run() called!")

	db := j.Configuration.DbConnection()
	updatedState, err := CasUpdateProcessingState(&j, Processing, db)
	if err != nil {
		j.recordProcessingError(err)
		return
	}

	if !updatedState {
		logg.LogTo("TRAINING_JOB", "%+v already processed.  Ignoring.", j)
		return
	}

	j.StdOutUrl = j.getStdOutCbfsUrl()
	j.StdErrUrl = j.getStdErrCbfsUrl()

	if err := j.extractData(); err != nil {
		j.recordProcessingError(err)
		return
	}

	if err := j.runCaffe(); err != nil {
		j.recordProcessingError(err)
		return
	}

	j.FinishedSuccessfully(j.Configuration.DbConnection(), "")

}

func (j *TrainingJob) GetProcessingState() ProcessingState {
	return j.ProcessingState
}

func (j *TrainingJob) SetProcessingState(newState ProcessingState) {
	j.ProcessingState = newState
}

func (j *TrainingJob) RefreshFromDB(db couch.Database) error {
	trainingJob := TrainingJob{}
	err := db.Retrieve(j.Id, &trainingJob)
	if err != nil {
		logg.LogTo("TRAINING_JOB", "Error getting latest: %v", err)
		return err
	}
	*j = trainingJob
	return nil
}

// call caffe train --solver=<work-dir>/spec.prototxt
func (j TrainingJob) runCaffe() error {

	logg.LogTo("TRAINING_JOB", "runCaffe()")

	// get the solver associated with this training job
	solver, err := j.getSolver()
	if err != nil {
		return fmt.Errorf("Error getting solver: %+v.  Err: %v", j, err)
	}

	// filename of solver prototxt, (ie, "solver.prototxt")
	_, solverFilename := filepath.Split(solver.SpecificationUrl)
	logg.LogTo("TRAINING_JOB", "solverFilename: %v", solverFilename)

	// build command args
	cmdArgs := []string{"train", fmt.Sprintf("--solver=%v", solverFilename)}
	caffePath := "caffe"

	// debugging
	logg.LogTo("TRAINING_JOB", "Running %v with args %v", caffePath, cmdArgs)
	logg.LogTo("TRAINING_JOB", "Path %v", os.Getenv("PATH"))
	out, _ := exec.Command("ls", "-alh", "/usr/local/bin").Output()
	logg.LogTo("TRAINING_JOB", "ls -alh /usr/local/bin: %v", string(out))

	// explicitly check if caffe binary found on the PATH
	path, err := exec.LookPath("caffe")
	if err != nil {
		logg.LogError(fmt.Errorf("caffe not found on path: %v", err))
	}
	logg.LogTo("TRAINING_JOB", "caffe found on path: %v", path)

	// Create Caffe command, but don't actually run it yet
	cmd := exec.Command(caffePath, cmdArgs...)

	// set the directory where the command will be run in (important
	// because we depend on relative file paths to work)
	cmd.Dir = j.getWorkDirectory()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Error running caffe: StdoutPipe(). Err: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Error running caffe: StderrPipe(). Err: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Error running caffe: cmd.Start(). Err: %v", err)
	}

	// read from stdout, stderr and write to temp files
	if err := j.saveCmdOutputToFiles(stdout, stderr); err != nil {
		return fmt.Errorf("Error running caffe: saveCmdOutput. Err: %v", err)
	}

	// wait for the command to complete
	runCommandErr := cmd.Wait()

	// read from temp files and write to cbfs.
	// initially I tried to write the stdout/stderr streams directly
	// to cbfs, but ran into an error related to the io.Seeker interface.
	if err := j.saveCmdOutputToCbfs(j.getStdOutPath()); err != nil {
		return fmt.Errorf("Error running caffe: could not save output to cbfs. Err: %v", err)
	}

	if err := j.saveCmdOutputToCbfs(j.getStdErrPath()); err != nil {
		return fmt.Errorf("Error running caffe: could not save output to cbfs. Err: %v", err)
	}

	// find out the name of the final model, eg snapshot_iter_200.caffemodel
	caffeModelFilename, err := j.getCaffeModelFilename()
	if err != nil {
		return fmt.Errorf("Error finding the caffe model file. Err: %v", err)
	}

	// upload caffemodel to cbfs as <training-job-id>/trained.caffemodel
	if err := j.uploadCaffeModelToCbfs(caffeModelFilename); err != nil {
		return fmt.Errorf("Error uploading caffe model to cbfs. Err: %v", err)
	}

	// update the training job to have the caffe model URL
	// set the url to the model, could be:
	//   relative (do this for now, convert to absolute later)
	//     - maybe cbfs/243224lkjlkj/caffe.model which a user can paste at end of API url
	//   absolute
	//     - http://host:8080/cbfs/243224lkjlkj/caffe.model
	//     - will need to be given public ip in config
	if err := j.updateCaffeModelUrl(); err != nil {
		return fmt.Errorf("Error updating caffe model url. Err: %v", err)
	}

	// TODO: add cbfs proxy so that we can get to this file
	// via http://host:8080/cbfs/243224lkjlkj/caffe.model

	return runCommandErr

}

func (j TrainingJob) uploadCaffeModelToCbfs(caffeModelFilename string) error {

	destPath := path.Join(j.Id, "trained.caffemodel")

	cbfs, err := j.Configuration.NewCbfsClient()
	if err != nil {
		return err
	}

	if err := saveFileToCbfs(caffeModelFilename, destPath, "application/octet-stream", cbfs); err != nil {
		return err
	}

	return nil
}

func (j TrainingJob) updateCaffeModelUrl() error {
	return nil
}

func (j TrainingJob) getCaffeModelFilename() (string, error) {

	// get the solver associated with this training job
	solver, err := j.getSolver()
	if err != nil {
		return "", fmt.Errorf("Error getting solver: %+v.  Err: %v", j, err)
	}

	// read into object with protobuf (must have already generated go protobuf code)
	solverParam, err := solver.getSolverParameter()
	if err != nil {
		return "", fmt.Errorf("Error getting solverParam. Err: %v", err)
	}

	maxIter := *solverParam.MaxIter
	snapshotPrefix := *solverParam.SnapshotPrefix

	// eg, snapshot_iter_200.caffemodel
	caffeModelFilename := fmt.Sprintf("%v_iter_%v.caffemodel", snapshotPrefix, maxIter)

	logg.LogTo("TRAINING_JOB", "caffeModelFilename: %v", caffeModelFilename)

	return caffeModelFilename, nil

}

func (j TrainingJob) getStdOutPath() string {
	return path.Join(j.getWorkDirectory(), "stdout")
}

func (j TrainingJob) getStdErrPath() string {
	return path.Join(j.getWorkDirectory(), "stderr")
}

func (j TrainingJob) getStdOutCbfsUrl() string {
	return fmt.Sprintf("%v/%v/%v", CBFS_URI_PREFIX, j.Id, path.Base(j.getStdOutPath()))
}

func (j TrainingJob) getStdErrCbfsUrl() string {
	return fmt.Sprintf("%v/%v/%v", CBFS_URI_PREFIX, j.Id, path.Base(j.getStdErrPath()))
}

func (j TrainingJob) saveCmdOutputToCbfs(sourcePath string) error {

	base := path.Base(sourcePath)
	destPath := fmt.Sprintf("%v/%v", j.Id, base)

	// todo: refactor to use saveFileToCbfs

	cbfs, err := cbfsclient.New(j.Configuration.CbfsUrl)
	if err != nil {
		return err
	}
	options := cbfsclient.PutOptions{
		ContentType: "text/plain",
	}

	logg.LogTo("TRAINING_JOB", "save to  destPath: %v", destPath)
	f, err := os.Open(sourcePath)
	if err != nil {
		return err
	}

	r := bufio.NewReader(f)

	if err := cbfs.Put("", destPath, r, options); err != nil {
		return fmt.Errorf("Error writing %v to cbfs: %v", destPath, err)
	}
	logg.LogTo("TRAINING_JOB", "Wrote %v to cbfs", destPath)
	return nil

}

func (j TrainingJob) saveCmdOutputToFiles(cmdStdout, cmdStderr io.ReadCloser) error {

	stdOutDoneChan := make(chan error, 1)
	stdErrDoneChan := make(chan error, 1)

	// also, Tee everything to this processes' stdout/stderr
	cmdStderrTee := io.TeeReader(cmdStderr, os.Stderr)
	cmdStdoutTee := io.TeeReader(cmdStdout, os.Stdout)

	// spawn goroutines to read from stdout/stderr
	go func() {
		if err := streamToFile(cmdStdoutTee, j.getStdOutPath()); err != nil {
			stdOutDoneChan <- err
		} else {
			stdOutDoneChan <- nil
		}

	}()

	go func() {
		if err := streamToFile(cmdStderrTee, j.getStdErrPath()); err != nil {
			stdErrDoneChan <- err
		} else {
			stdErrDoneChan <- nil
		}

	}()

	// wait for goroutines
	stdOutResult := <-stdOutDoneChan
	stdErrResult := <-stdErrDoneChan

	// check for errors
	results := []error{stdOutResult, stdErrResult}
	for _, result := range results {
		if result != nil {
			return fmt.Errorf("Saving cmd output failed: %v", result)
		}
	}

	return nil
}

func (j TrainingJob) extractData() error {

	// get the solver associated with this training job
	solver, err := j.getSolver()
	if err != nil {
		return fmt.Errorf("Error getting solver: %+v.  Err: %v", j, err)
	}

	// create a work directory based on config, eg, /usr/lib/elasticthought/<job-id>
	if err := j.createWorkDirectory(); err != nil {
		return fmt.Errorf("Error creating work dir: %+v.  Err: %v", j, err)
	}

	// read prototext from cbfs, write to work dir
	if err := j.writeSpecToFile(*solver); err != nil {
		return fmt.Errorf("Error saving specifcation: %+v.  Err: %v", j, err)
	}

	// download and untar the training and test .tar.gz files associated w/ solver
	if err := j.saveTrainTestData(*solver); err != nil {
		return fmt.Errorf("Error saving train/test data: %+v.  Err: %v", j, err)
	}

	return nil

}

func (j TrainingJob) saveTrainTestData(s Solver) error {

	if err := s.SaveTrainTestData(j.Configuration, j.getWorkDirectory()); err != nil {
		return err
	}
	return nil

}

// Codereview: de-dupe
func (j TrainingJob) recordProcessingError(err error) {
	logg.LogError(err)
	db := j.Configuration.DbConnection()
	if err := j.Failed(db, err); err != nil {
		errMsg := fmt.Errorf("Error setting training job as failed: %v", err)
		logg.LogError(errMsg)
	}
}

func (j TrainingJob) getWorkDirectory() string {
	return filepath.Join(j.Configuration.WorkDirectory, j.Id)
}

func (j TrainingJob) createWorkDirectory() error {
	workDir := j.getWorkDirectory()
	logg.LogTo("TRAINING_JOB", "Creating dir: %v", workDir)
	return Mkdir(workDir)
}

func (j TrainingJob) getSolver() (*Solver, error) {
	db := j.Configuration.DbConnection()
	solver := &Solver{}
	err := db.Retrieve(j.SolverId, solver)
	if err != nil {
		errMsg := fmt.Errorf("Didn't retrieve: %v - %v", j.SolverId, err)
		logg.LogError(errMsg)
		return nil, errMsg
	}
	solver.Configuration = j.Configuration
	return solver, nil
}

func (j TrainingJob) writeSpecToFile(s Solver) error {

	if err := s.writeSpecToFile(j.Configuration, j.getWorkDirectory()); err != nil {
		return err
	}
	logg.LogTo("TRAINING_JOB", "Saved specification: %v", j.getWorkDirectory())
	return nil

}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
// Codereview: de-dupe
func (j TrainingJob) Insert(db couch.Database) (*TrainingJob, error) {

	id, _, err := db.Insert(j)
	if err != nil {
		err := fmt.Errorf("Error inserting training job: %+v.  Err: %v", j, err)
		return nil, err
	}

	// load dataset object from db (so we have id/rev fields)
	trainingJob := &TrainingJob{}
	err = db.Retrieve(id, trainingJob)
	if err != nil {
		err := fmt.Errorf("Error fetching training job: %v.  Err: %v", id, err)
		return nil, err
	}

	return trainingJob, nil

}

// Update the state to record that it failed
// Codereview: de-dupe
func (j TrainingJob) Failed(db couch.Database, processingErr error) error {

	_, err := CasUpdateProcessingState(&j, Failed, db)
	if err != nil {
		return err
	}

	logg.LogTo("TRAINING_JOB", "updating processing log")

	j.ProcessingLog = fmt.Sprintf("%v", processingErr)

	// TODO: retry if 409 error
	_, err = db.Edit(j)

	if err != nil {
		return err
	}

	return nil

}

// Update the state to record that it succeeded
// Codereview: de-dupe
func (j TrainingJob) FinishedSuccessfully(db couch.Database, logPath string) error {

	_, err := CasUpdateProcessingState(&j, FinishedSuccessfully, db)
	if err != nil {
		return err
	}

	j.ProcessingLog = logPath

	// TODO: retry if 409 error
	_, err = db.Edit(j)

	if err != nil {
		return err
	}

	return nil

}
