package elasticthought

import (
	"encoding/json"
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

	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer(n.Configuration.NsqdTopic, channelName, config)
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {

		logg.LogTo("NSQ_WORKER", "Got a message!: %v", string(message.Body))

		// create jobDescriptor from json
		jobDescriptor := JobDescriptor{}
		err := json.Unmarshal(message.Body, &jobDescriptor)
		if err != nil {
			bodyStr := string(message.Body)
			logg.LogTo("NSQ_WORKER", "Error unmarshalling msg: %v", bodyStr)
			return err
		}

		logg.LogTo("NSQ_WORKER", "Job descriptor: %+v", jobDescriptor)

		// create job from job descriptor
		job, err := CreateJob(n.Configuration, jobDescriptor)
		if err != nil {
			logg.LogTo("NSQ_WORKER", "Error creating job from: %+v", jobDescriptor)
			return err
		}

		logg.LogTo("NSQ_WORKER", "Job: %+v", job)

		// run job
		go job.Run(nil)

		return nil
	}))
	err := q.ConnectToNSQLookupd(n.Configuration.NsqLookupdUrl)
	if err != nil {
		errMsg := fmt.Errorf("Error connecting to nsq: %v", err)
		logg.LogError(errMsg)
	}

	logg.LogTo("NSQ_WORKER", "connected to nsq as a consumer")

}
