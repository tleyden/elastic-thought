package elasticthought

type NsqJobRunner struct {
	Configuration Configuration
}

func NewNsqJobRunner(c Configuration) *NsqJobRunner {
	return &NsqJobRunner{
		Configuration: c,
	}
}

func (j NsqJobRunner) ScheduleJob(jobDescriptor JobDescriptor) error {

	// connect to nsq

	// serialize job descriptor to json

	// publish to nsq

	return nil
}
