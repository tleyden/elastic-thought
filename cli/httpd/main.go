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

	if err := et.EnvironmentSanityCheck(config); err != nil {
		logg.LogFatal("Failed environment sanity check: %v", err)
		return
	}

	var jobScheduler et.JobScheduler

	switch config.QueueType {
	case et.Nsq:
		jobScheduler = et.NewNsqJobScheduler(config)
	case et.Goroutine:
		jobScheduler = et.NewInProcessJobScheduler(config)
	default:
		logg.LogFatal("Unexpected queue type: %v", config.QueueType)
	}

	context := &et.EndpointContext{
		Configuration: config,
	}

	changesListener, err := et.NewChangesListener(config, jobScheduler)
	if err != nil {
		logg.LogPanic("Error creating changes listener: %v", err)
	}
	go changesListener.FollowChangesFeed()

	r := gin.Default()
	r.Use(et.DbConnector(config.DbUrl))
	r.POST("/users", context.CreateUserEndpoint)

	authorized := r.Group("/")
	authorized.Use(et.DbAuthRequired())
	{
		authorized.POST("/datafiles", context.CreateDataFileEndpoint)
		authorized.POST("/datasets", context.CreateDataSetsEndpoint)
		authorized.POST("/solvers", context.CreateSolverEndpoint)
		authorized.POST("/training-jobs", context.CreateTrainingJob)
		authorized.POST("/classifiers", context.CreateClassifierEndpoint)
		authorized.POST("/classifiers/:classifier-id/classify", context.CreateClassificationJobEndpoint)
	}

	// Listen and serve on 0.0.0.0:8080
	r.Run(":8080")

}
