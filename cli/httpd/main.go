// Command line utility to launch the ElasticThought REST API server.
package main

import (
	"github.com/couchbaselabs/logg"
	"github.com/dustin/go-couch"
	"github.com/gin-gonic/gin"
	et "github.com/tleyden/elastic-thought"
)

func init() {
	logg.LogKeys["CLI"] = true
	logg.LogKeys["REST"] = true
}

func main() {

	r := gin.Default()
	r.Use(et.DbConnector("http://localhost:4985/elasticthought"))

	r.GET("/ping", func(c *gin.Context) {
		db := c.MustGet("db").(couch.Database)
		c.String(200, "db is: %+v", db)
	})

	r.POST("/users", et.CreateUserEndpoint)
	authorized := r.Group("/")
	authorized.Use(et.DbAuthRequired())
	{
		authorized.POST("/datafiles", et.CreateDataFileEndpoint)
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
