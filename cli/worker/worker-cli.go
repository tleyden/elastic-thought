// Command line utility to launch an ElasticThought worker
package main

import (
	"github.com/couchbaselabs/logg"
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

	worker := NewNsqWorker(config)
	go worker.HandleEvents()

	select {} // block forever

}
