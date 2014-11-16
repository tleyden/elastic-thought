package elasticthought

import "github.com/couchbaselabs/logg"

// Run worker jobs in a goroutine in the rest server process (as oppposed to using nsq)
// Makes certain testing easier
type InProcessJobScheduler struct {
	Configuration Configuration
}

func NewInProcessJobScheduler(c Configuration) *InProcessJobScheduler {
	return &InProcessJobScheduler{
		Configuration: c,
	}
}

func (j InProcessJobScheduler) ScheduleJob(jobDescriptor JobDescriptor) error {

	// create job locally and fire off go-routine
	logg.Log("in process job runner called with: %+v", jobDescriptor)

	job, err := CreateJob(j.Configuration, jobDescriptor)
	if err != nil {
		return err
	}

	go job.Run()

	return nil
}
