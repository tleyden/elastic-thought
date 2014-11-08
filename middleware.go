package elasticthought

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/couchbaselabs/logg"
	"github.com/dustin/go-couch"
	"github.com/gin-gonic/gin"
)

func DbConnector(dbUrl string) gin.HandlerFunc {

	return func(c *gin.Context) {

		db, err := couch.Connect(dbUrl)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error %v | dbUrl: %v", err, dbUrl))
			logg.LogError(err)
			c.Fail(500, err)
		}

		c.Set("db", db)

		c.Next()

	}

}

func DbAuthRequired() gin.HandlerFunc {

	return func(c *gin.Context) {

		db := c.MustGet("db").(couch.Database)

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

		c.Set("user", *user)

		c.Next()

	}

}
