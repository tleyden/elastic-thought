// Command line utility to launch an ElasticThought worker
package main

import (
	"github.com/couchbaselabs/logg"
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

	worker := et.NewNsqWorker(config)
	go worker.HandleEvents()

	select {} // block forever

}
