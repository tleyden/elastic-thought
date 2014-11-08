package elasticthought

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/couchbaselabs/logg"
	"github.com/dustin/go-couch"
	"github.com/gin-gonic/gin"
)

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

func CreateDataFileEndpoint(c *gin.Context) {

	user := c.MustGet("user").(User)

	// create a new Datfile object

	// _changes listener will see it and process it (download and save to s3)

	// return uuid of Dataafile object

	c.String(200, "user is: %+v", user)
}
