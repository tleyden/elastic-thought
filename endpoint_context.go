package elasticthought

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/couchbaselabs/cbfs/client"
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

	datafile := NewDatafile()
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
	id, _, err := db.Insert(datafile)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new datafile: %v", err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	c.JSON(201, gin.H{"id": id})

}

// Creates datasets from a datafile
func (e EndpointContext) CreateDataSetsEndpoint(c *gin.Context) {

	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)
	logg.LogTo("REST", "user: %v db: %v", user, db)

	dataset := NewDataset()

	// bind the input struct to the JSON request
	if ok := c.Bind(dataset); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	// save dataset in db
	dataset, err := dataset.Insert(db)
	if err != nil {
		c.Fail(500, err)
		return
	}

	// update with urls of training/testing artifacts (which don't exist yet)
	dataset, err = dataset.AddArtifactUrls(db)
	if err != nil {
		errMsg := fmt.Sprintf("Error updating dataset: %v.  Err: %v", dataset.Id, err)
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

	solver := NewSolver()

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
	cbfs, err := cbfsclient.New(e.Configuration.CbfsUrl)
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		c.Fail(500, errMsg)
		return
	}
	logg.LogTo("REST", "cbfs: %+v", cbfs)

	// download contents of specification-url into cbfs://<solver-id>/spec.prototxt
	// and update solver object's specification-url with cbfs url.
	// ditto for specification-net-url
	solver, err = solver.SaveSpec(db, cbfs)
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

	trainingJob := NewTrainingJob()
	trainingJob.UserID = user.Id

	// bind the input struct to the JSON request
	if ok := c.Bind(trainingJob); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "trainingJob: %+v", trainingJob)

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
