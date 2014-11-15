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

	var decodedJson struct {
		DatafileId string `json:"datafile-id" binding:"required"`
		Split      struct {
			Training float32 `json:"training" binding:"required"`
			Testing  float32 `json:"testing" binding:"required"`
		} `json:"split" binding:"required"`
	}

	// bind the input struct to the JSON request
	if ok := c.Bind(&decodedJson); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	// get Datafile object in db

	// create two new Dataset objects that reference this Datafile

	// for each new dataset object, add
	// add message to queue so that a worker processes it
	// config := nsq.NewConfig()

	// return
	/*

	   {
	       "datasets": [
	           {
	               "datafile-id": "datafile-uuid",
	               "id": "training-dataset-uuid",
	               "name":"training",
	               "split-percentage": 0.7
	           },
	           {
	               "datafile-id": "datafile-uuid",
	               "id": "testing-dataset-uuid",
	               "name":"testing",
	               "split-percentage": 0.3
	           }
	       ]
	   }

	*/

	// c.String(200, "input is: %+v", input2)

}
