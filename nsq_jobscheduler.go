package elasticthought

type NsqJobScheduler struct {
	Configuration Configuration
}

func NewNsqJobScheduler(c Configuration) *NsqJobScheduler {
	return &NsqJobScheduler{
		Configuration: c,
	}
}

func (j NsqJobScheduler) ScheduleJob(jobDescriptor JobDescriptor) error {

	// connect to nsq

	// serialize job descriptor to json

	// publish to nsq

	return nil
}
