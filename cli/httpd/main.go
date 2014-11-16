// Command line utility to launch the ElasticThought REST API server.
package main

import (
	"github.com/couchbaselabs/logg"
	"github.com/gin-gonic/gin"
	et "github.com/tleyden/elastic-thought"
)

func init() {
	et.EnableAllLogKeys()
}

func main() {

	config := *(et.NewDefaultConfiguration()) // TODO: get these vals from cmd line args

	// TODO: make this a config to choose either the in process job runner
	// or an NSQJobRunner
	// jobScheduler := et.NewInProcessJobScheduler(config)
	jobScheduler := et.NewNsqJobScheduler(config)

	changesListener, err := et.NewChangesListener(config, jobScheduler)
	if err != nil {
		logg.LogPanic("Error creating changes listener: %v", err)
	}
	go changesListener.FollowChangesFeed()

	r := gin.Default()
	r.Use(et.DbConnector(config.DbUrl))
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
