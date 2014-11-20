package elasticthought

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/couchbaselabs/logg"
	"github.com/gin-gonic/gin"
	"github.com/tleyden/go-couch"
)

// Creates a new user
func CreateUserEndpoint(c *gin.Context) {

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
func CreateDataFileEndpoint(c *gin.Context) {

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
func CreateDataSetsEndpoint(c *gin.Context) {

	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)
	logg.LogTo("REST", "user: %v db: %v", user, db)

	datasetInput := NewDataset()

	// bind the input struct to the JSON request
	if ok := c.Bind(datasetInput); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	// save dataset object in db -- it will get picked up and processed
	// by changes listener
	id, _, err := db.Insert(datasetInput)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new dataset: %v", err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	// load dataset object from db (so we have id/rev fields)
	dataset := Dataset{}
	err = db.Retrieve(id, &dataset)
	if err != nil {
		errMsg := fmt.Sprintf("Error fetching dataset w/ id: %v.  Err: %v", id, err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	// update with urls of training/testing artifacts (which don't exist yet)
	dataset, err = dataset.AddArtifactUrls(db)
	if err != nil {
		errMsg := fmt.Sprintf("Error updating dataset: %v.  Err: %v", id, err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	c.JSON(201, dataset)

}

// Creates a solver
func CreateSolverEndpoint(c *gin.Context) {

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

	// download contents of solver-spec-url into cbfs://<solver-id>/spec.prototxt

	// update solver object's solver-spec-url  with cbfs url

	// return solver object
	c.JSON(201, *solver)

}
