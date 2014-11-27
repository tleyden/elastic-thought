// Command line utility to launch an ElasticThought worker
package main

import (
	"fmt"

	"github.com/couchbaselabs/logg"
	et "github.com/tleyden/elastic-thought"
)

func init() {
	et.EnableAllLogKeys()
}

func main() {

	config := *(et.NewDefaultConfiguration()) // TODO: get these vals from cmd line args

	if err := et.EnvironmentSanityCheck(config); err != nil {
		logg.LogError(fmt.Errorf("Failed environment sanity check: %v", err))
	}

	worker := et.NewNsqWorker(config)
	go worker.HandleEvents()

	select {} // block forever

}
