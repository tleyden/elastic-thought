// Command line utility to launch an ElasticThought worker
package main

import et "github.com/tleyden/elastic-thought"

func init() {
	et.EnableAllLogKeys()
}

func main() {

	config := *(et.NewDefaultConfiguration()) // TODO: get these vals from cmd line args

	worker := et.NewNsqWorker(config)
	go worker.HandleEvents()

	select {} // block forever

}
