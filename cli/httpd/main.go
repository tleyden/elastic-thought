// Command line utility to launch the ElasticThought REST API server.
package main

import (
	"github.com/couchbaselabs/logg"
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

	r.POST("/users", et.CreateUserEndpoint)

	authorized := r.Group("/")
	authorized.Use(et.DbAuthRequired())
	{
		authorized.POST("/datafiles", et.CreateDataFileEndpoint)
		authorized.POST("/datasets", et.CreateDataSetsEndpoint)
	}

	// Listen and serve on 0.0.0.0:8080
	r.Run(":8080")

}
