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
	logg.LogKeys["DATASET_SPLITTER"] = true
}

func main() {

	config := et.Configuration{}
	config.DbUrl = "http://localhost:4985/elasticthought"
	config.NsqLookupdUrl = "127.0.0.1:4161"
	config.NsqdUrl = "127.0.0.1:4150"
	config.NsqdTopic = "elastic-thought"

	// TODO: make this a config to choose either the in process job runner
	// or an NSQJobRunner
	jobRunner := et.NewInProcessJobScheduler(config)

	changesListener, err := et.NewChangesListener(config, jobRunner)
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
