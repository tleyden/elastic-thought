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
	id, rev, err := db.InsertWith(newUser, newUser.DocId())
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new user: %v", err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	c.String(201, "Created new user with id: %v rev: %v", id, rev)

}

// Creates a datafile
func CreateDataFileEndpoint(c *gin.Context) {

	user := c.MustGet("user").(User)
	db := c.MustGet("db").(couch.Database)

	datafile := &Datafile{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_DATAFILE},
		UserID:            user.DocId(),
	}

	if ok := c.Bind(&datafile); !ok {
		errMsg := fmt.Sprintf("Invalid datafile")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "datafile: %+v", datafile)

	// create a new Datafile object in db

	id, rev, err := db.Insert(datafile)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new datafile: %v", err)
		c.Fail(500, errors.New(errMsg))
		return
	}

	// return uuid of Dataafile object

	c.String(201, "created datafile id: %v rev: %v", id, rev)

}

// Creates datasets from a datafile
func CreateDataSetsEndpoint(c *gin.Context) {

	user := c.MustGet("user").(User)

	// create a new Datafile object in db

	// create two new Dataset objects that reference this Datafile

	// _changes listener will see new Dataset objects and process them

	// return uuid of Dataafile object

	c.String(200, "user is: %+v", user)

}
