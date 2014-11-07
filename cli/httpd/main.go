// Command line utility to launch the ElasticThought REST API server.
package main

import (
	"fmt"
	"net/http"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/elastic-thought"
)

func init() {
	logg.LogKeys["CLI"] = true
}

func main() {

	dbUrl := "http://localhost:4985/elasticthought" // TODO: cli param
	port := 8080                                    // TODO: cli-param

	restApiServer := elasticthought.NewRestApiServer(dbUrl)

	logg.LogTo("CLI", "Starting webserver on port: %v", port)

	listenPort := fmt.Sprintf(":%v", port)
	logg.LogError(http.ListenAndServe(listenPort, restApiServer.RestApiRouter()))

}
