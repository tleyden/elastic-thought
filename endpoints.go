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

	db := c.MustGet("db").(couch.Database)

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

	user := c.MustGet("user").(User)
	db := c.MustGet("db").(couch.Database)

	datafile := &Datafile{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_DATAFILE},
		UserID:            user.DocId(),
	}

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

type datasetsInput struct {
	DatafileId string `json:"datafile-id" binding:"required"`
	Split      struct {
		Training float32 `json:"training" binding:"required"`
		Testing  float32 `json:"testing" binding:"required"`
	} `json:"split" binding:"required"`
}

// Creates datasets from a datafile
func CreateDataSetsEndpoint(c *gin.Context) {

	user := c.MustGet("user").(User)
	db := c.MustGet("db").(couch.Database)
	logg.LogTo("REST", "user: %v db: %v", user, db)

	input := datasetsInput{}

	// bind the input struct to the JSON request
	if ok := c.Bind(&input); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	// get Datafile object in db

	// create two new Dataset objects that reference this Datafile

	// add message to queue so that a worker processes it
	// config := nsq.NewConfig()

	// return uuid of Datafile object

	c.String(200, "input is: %+v", input)

}
