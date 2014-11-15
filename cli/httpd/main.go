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
	logg.LogKeys["CHANGES"] = true
}

func main() {

	dbUrl := "http://localhost:4985/elasticthought"

	changesListener, err := et.NewChangesListener(dbUrl)
	if err != nil {
		logg.LogPanic("Error creating changes listener: %v", err)
	}
	go changesListener.FollowChangesFeed()

	r := gin.Default()
	r.Use(et.DbConnector(dbUrl))
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
