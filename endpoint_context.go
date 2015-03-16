package elasticthought

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/couchbaselabs/logg"
	"github.com/gin-gonic/gin"
	"github.com/tleyden/go-couch"
)

type EndpointContext struct {
	Configuration Configuration
}

// Creates a new user
func (e EndpointContext) CreateUserEndpoint(c *gin.Context) {

	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)

	// parse in a user object from the POST request
	decoder := json.NewDecoder(c.Request.Body)
	userToCreate := NewUser()
	err := decoder.Decode(userToCreate)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to parse user params: %v", err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	// make sure this user isn't already in the db
	existingUser := NewUser()
	err = db.Retrieve(userToCreate.DocId(), existingUser)
	if err == nil {
		errMsg := fmt.Sprintf("User already exists: %+v", *existingUser)
		c.Fail(500, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "Did not find existing user, ok to create")

	// create a new user and return 201
	newUser := NewUserFromUser(*userToCreate)
	_, _, err = db.InsertWith(newUser, newUser.DocId())
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new user: %v", err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	c.String(201, "")

}

// Creates a datafile
func (e EndpointContext) CreateDataFileEndpoint(c *gin.Context) {

	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)

	datafile := NewDatafile(e.Configuration)
	datafile.UserID = user.DocId()

	// bind the Datafile to the JSON request, which will bind the
	// url field or throw an error.
	if ok := c.Bind(&datafile); !ok {
		errMsg := fmt.Sprintf("Invalid datafile")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "datafile: %+v", datafile)

	// create a new Datafile object in db
	datafile, err := datafile.Save(db)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new datafile: %v", err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	c.JSON(201, gin.H{"id": datafile.Id})

}

// Creates datasets from a datafile
func (e EndpointContext) CreateDataSetsEndpoint(c *gin.Context) {

	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)
	logg.LogTo("REST", "user: %v db: %v", user, db)

	dataset := NewDataset(e.Configuration)

	// bind the input struct to the JSON request
	if ok := c.Bind(dataset); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "dataset: %+v", dataset)

	// save dataset in db
	if err := dataset.Insert(); err != nil {
		c.Fail(500, err)
		return
	}

	// the changes listener will see new datafile and download to cbfs

	// update with urls of training/testing artifacts (which don't exist yet)
	if err := dataset.AddArtifactUrls(); err != nil {
		errMsg := fmt.Sprintf("Error updating dataset: %+v.  Err: %v", dataset, err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	c.JSON(201, dataset)

}

// Creates a solver
func (e EndpointContext) CreateSolverEndpoint(c *gin.Context) {

	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)
	logg.LogTo("REST", "user: %v db: %v", user, db)

	solver := NewSolver(e.Configuration)

	// bind the input struct to the JSON request
	if ok := c.Bind(solver); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "solver: %+v", solver)

	// save solver in db
	solver, err := solver.Insert(db)
	if err != nil {
		c.Fail(500, err)
		return
	}

	// Create a cbfs client
	cbfs, err := NewBlobStore(e.Configuration.CbfsUrl)
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		c.Fail(500, errMsg)
		return
	}
	logg.LogTo("REST", "cbfs: %+v", cbfs)

	// download contents of specification-url into cbfs://<solver-id>/spec.prototxt
	// and update solver object's specification-url with cbfs url.
	// ditto for specification-net-url
	solver, err = solver.DownloadSpecToBlobStore(db, cbfs)
	if err != nil {
		c.Fail(500, err)
		return
	}

	// return solver object
	c.JSON(201, *solver)

}

// Create Training Job
func (e EndpointContext) CreateTrainingJob(c *gin.Context) {

	// bind to json
	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)

	trainingJob := NewTrainingJob(e.Configuration)
	trainingJob.UserID = user.Id

	// bind the input struct to the JSON request
	if ok := c.Bind(trainingJob); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "Create new TrainingJob: %+v", trainingJob)

	// save training job in db
	trainingJob, err := trainingJob.Insert(db)
	if err != nil {
		c.Fail(500, err)
		return
	}

	// job will get kicked off by changes listener

	// return solver object
	c.JSON(201, *trainingJob)

}

// Creates a classifier
func (e EndpointContext) CreateClassifierEndpoint(c *gin.Context) {

	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)
	logg.LogTo("REST", "user: %v db: %v", user, db)

	classifier := NewClassifier(e.Configuration)

	// bind the input struct to the JSON request
	if ok := c.Bind(classifier); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "classifier: %+v", classifier)

	// make sure the classifier points to a valid training job
	logg.LogTo("REST", "Validating classifier")
	if err := classifier.Validate(); err != nil {
		logg.LogTo("REST", "Classifier failed validation: %v", err)
		c.Fail(400, err)
		return
	}

	// save classifier in db
	logg.LogTo("REST", "Save classifier to db")
	err := classifier.Insert()
	if err != nil {
		c.Fail(500, err)
		return
	}

	// Create a cbfs client
	cbfs, err := NewBlobStore(e.Configuration.CbfsUrl)
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		c.Fail(500, errMsg)
		return
	}

	// download contents of spec-url into cbfs://<classifier-id>/classifier.prototxt
	logg.LogTo("REST", "Save classifier.prototxt %v to cbfs", classifier.SpecificationUrl)
	destPath := path.Join(classifier.Id, "classifier.prototxt")
	if err := saveUrlToBlobStore(classifier.SpecificationUrl, destPath, cbfs); err != nil {
		c.Fail(500, err)
		return

	}

	// update the spec url to point to the classifier.prototxt in cbfs
	logg.LogTo("REST", "update the spec url to point to the classifier.prototxt in cbfs")
	specUrlCbfs := fmt.Sprintf("%v%v", CBFS_URI_PREFIX, destPath)
	if err := classifier.SetSpecificationUrl(specUrlCbfs); err != nil {
		c.Fail(500, err)
		return
	}

	// return classifier object
	c.JSON(201, *classifier)

}

func (e EndpointContext) CreateClassificationJobEndpoint(c *gin.Context) {

	_ = c.MustGet(MIDDLEWARE_KEY_USER).(User)
	_ = c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)

	classifierId := c.Params.ByName("classifier-id")

	classifier := NewClassifier(e.Configuration)
	if err := classifier.Find(classifierId); err != nil {
		err = fmt.Errorf("Unable to find classifier with id: %v.  Err: %v", classifierId, err)
		c.Fail(500, err)
		return
	}

	// create a new classifier job
	classifyJob := NewClassifyJob(e.Configuration)
	classifyJob.ClassifierID = classifier.Id

	request := c.Request
	err := request.ParseMultipartForm(100000000) // ~100 MB
	if err != nil {
		c.Fail(500, err)
		return
	}

	// TODO: currently ignores file upload files

	// get the form values with the image urls
	multipartForm := request.MultipartForm
	urls := multipartForm.Value["urls"]

	// manually create a new uuid here so we can refer to the id
	// before persisting the object to the db
	classifyJobId := NewUuid()
	classifyJob.Id = classifyJobId

	// add each image to cbfs
	emptyResults := map[string]string{}
	for _, url := range urls {

		hash := sha1.Sum([]byte(url))
		hashHexString := fmt.Sprintf("% x", hash)
		hashHexString = strings.Replace(hashHexString, " ", "", -1)

		// save image url to cbfs
		dest := path.Join(classifyJob.Id, hashHexString)

		cbfsclient, err := e.Configuration.NewBlobStoreClient()
		if err != nil {
			c.Fail(500, err)
			return
		}

		if err := saveUrlToBlobStore(url, dest, cbfsclient); err != nil {
			c.Fail(500, err)
			return
		}

		imageUrlCbfs := fmt.Sprintf("%v%v", CBFS_URI_PREFIX, dest)

		emptyResults[imageUrlCbfs] = "pending"

	}

	classifyJob.Results = emptyResults

	if err := classifyJob.Insert(); err != nil {
		c.Fail(500, err)
		return
	}

	// changes listener will see job and kick off processing

	// return classifier object
	c.JSON(201, *classifyJob)

}
