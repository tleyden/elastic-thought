// Command line utility to launch an ElasticThought worker
package main

import et "github.com/tleyden/elastic-thought"

func init() {
	et.EnableAllLogKeys()
}

func main() {

	config := et.Configuration{}
	config.DbUrl = "http://localhost:4985/elasticthought"
	config.NsqLookupdUrl = "127.0.0.1:4161"
	config.NsqdUrl = "127.0.0.1:4150"
	config.NsqdTopic = "elastic-thought"

	worker := et.NewNsqWorker(config)
	go worker.HandleEvents()

	select {} // block forever

}
