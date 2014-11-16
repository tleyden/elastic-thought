package elasticthought

import (
	"fmt"

	"github.com/bitly/go-nsq"
	"github.com/couchbaselabs/logg"
)

// A worker which pulls jobs off of NSQ and processes them
type NsqWorker struct {
	Configuration Configuration
}

func NewNsqWorker(c Configuration) *NsqWorker {
	return &NsqWorker{
		Configuration: c,
	}
}

func (n NsqWorker) HandleEvents() {

	// not really sure if I need to use channels at all here,
	// since at the moment there is only one worker
	channelName := "channel"

	// pull event off of nsql topic

	// create jobDescriptor from json

	// create job from job descriptor
	// job := CreateJob(n.Configuration, jobDescriptor)

	// run job
	// go job.Run()

	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer(n.Configuration.NsqdTopic, channelName, config)
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		logg.LogTo("NSQ_WORKER", "Got a message!!!: %v", message)
		return nil
	}))
	err := q.ConnectToNSQLookupd(n.Configuration.NsqLookupdUrl)
	if err != nil {
		errMsg := fmt.Errorf("Error connecting to nsq: %v", err)
		logg.LogError(errMsg)
	}

	logg.LogTo("NSQ_WORKER", "connected to nsq as a consumer")

}
