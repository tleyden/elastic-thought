// Command line utility to launch the ElasticThought REST API server.
package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/couchbaselabs/logg"
	"github.com/dustin/go-couch"
	"github.com/gin-gonic/gin"
)

func init() {
	logg.LogKeys["CLI"] = true
	logg.LogKeys["REST"] = true
}

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

		c.Set("user", user)
		c.Set("password", password)

		c.Next()

	}

}

func main() {

	r := gin.Default()
	r.Use(DbConnector("http://localhost:4985/elasticthought"))

	r.GET("/ping", func(c *gin.Context) {
		db := c.MustGet("db").(couch.Database)
		c.String(200, "db is: %+v", db)
	})

	datafilesEndpoint := func(c *gin.Context) {
		user := c.MustGet("user").(string)
		c.String(200, "user is: %+v", user)
	}

	authorized := r.Group("/")
	authorized.Use(DbAuthRequired())
	{
		authorized.POST("/datafiles", datafilesEndpoint)
	}

	// Listen and server on 0.0.0.0:8080
	r.Run(":8080")

	/*
		dbUrl := "http://localhost:4985/elasticthought" // TODO: cli param
		port := 8080                                    // TODO: cli-param

		restApiServer := elasticthought.NewRestApiServer(dbUrl)

		logg.LogTo("CLI", "Starting webserver on port: %v", port)

		listenPort := fmt.Sprintf(":%v", port)
		logg.LogError(http.ListenAndServe(listenPort, restApiServer.RestApiRouter()))
	*/

}
