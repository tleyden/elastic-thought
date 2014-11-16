package elasticthought

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

	for {
		// pull event off of nsql topic

		// create jobDescriptor from json

		// create job from job descriptor
		// job := CreateJob(n.Configuration, jobDescriptor)

		// run job
		// go job.Run()

	}

}
