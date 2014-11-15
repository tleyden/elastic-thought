package elasticthought

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/couchbaselabs/logg"
	"github.com/gin-gonic/gin"
	"github.com/tleyden/go-couch"
)

const (
	MIDDLEWARE_KEY_DB   = "db"
	MIDDLEWARE_KEY_USER = "user"
)

// Gin middleware to connnect to the Sync Gw database given in the
// dbUrl parameter, and set the connection object into the context.
// This creates a new connection for each request, which is ultra-conservative
// in case the connection object isn't safe to use among multiple goroutines
// (and I believe it is).  If it becomes a bottleneck, it's easy to create
// another middleware that re-uses an existing connection.
func DbConnector(dbUrl string) gin.HandlerFunc {

	return func(c *gin.Context) {

		// make sure the db url does not have a trailing slash
		if strings.HasSuffix(dbUrl, "/") {
			err := errors.New(fmt.Sprintf("dbUrl needs trailing slash: %v", dbUrl))
			logg.LogError(err)
			c.Fail(500, err)
			return
		}

		db, err := couch.Connect(dbUrl)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error %v | dbUrl: %v", err, dbUrl))
			logg.LogError(err)
			c.Fail(500, err)
			return
		}

		c.Set(MIDDLEWARE_KEY_DB, db)

		c.Next()

	}

}

// Gin middleware to authenticate the user specified in the Basic Auth
// Authorization header.  It will lookup the user in the database (this
// middleware requires the use of the DbConnector middleware to have run
// before it), and then add to the Gin Context.
func DbAuthRequired() gin.HandlerFunc {

	return func(c *gin.Context) {

		db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)

		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

		logg.LogTo("REST", "db: %v, auth: %v", db, auth)

		if len(auth) != 2 || auth[0] != "Basic" {
			err := errors.New("bad syntax in auth header")
			c.Fail(401, err)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 {
			err := errors.New("expected user:pass in auth header")
			c.Fail(401, err)
			return
		}

		username := pair[0]
		password := pair[1]

		logg.LogTo("REST", "user: %v pass: %v", username, password)

		user, err := AuthenticateUser(db, username, password)
		if err != nil {
			msg := fmt.Sprintf("Failed to authenticate user in DB: %v", err)
			err := errors.New(msg)
			c.Fail(401, err)
			return
		}

		c.Set(MIDDLEWARE_KEY_USER, *user)

		c.Next()

	}

}
