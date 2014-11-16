package elasticthought

import (
	"encoding/json"

	"github.com/bitly/go-nsq"
	"github.com/couchbaselabs/logg"
)

type NsqJobScheduler struct {
	Configuration Configuration
}

func NewNsqJobScheduler(c Configuration) *NsqJobScheduler {
	return &NsqJobScheduler{
		Configuration: c,
	}
}

func (j NsqJobScheduler) ScheduleJob(jobDescriptor JobDescriptor) error {

	config := nsq.NewConfig()
	w, _ := nsq.NewProducer(j.Configuration.NsqdUrl, config)

	data, err := json.Marshal(jobDescriptor)
	if err != nil {
		return err
	}

	err = w.Publish(j.Configuration.NsqdTopic, data)
	if err != nil {
		return err
	}

	logg.LogTo("JOB_SCHEDULER", "Published to nsq: %v", string(data))

	w.Stop()

	return nil
}
