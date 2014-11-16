package elasticthought

import "github.com/couchbaselabs/logg"

// TODO: rename to InProcessJobScheduler
type InProcessJobRunner struct {
	Configuration Configuration
}

func NewInProcessJobRunner(c Configuration) *InProcessJobRunner {
	return &InProcessJobRunner{
		Configuration: c,
	}
}

func (j InProcessJobRunner) ScheduleJob(jobDescriptor JobDescriptor) error {

	// create job locally and fire off go-routine
	logg.Log("in process job runner called with: %+v", jobDescriptor)

	job := CreateJob(jobDescriptor)
	go job.Run()

	return nil
}

func CreateJob(jobDescriptor JobDescriptor) Runnable {
	return DatasetSplitter{}
}
