package elasticthought

import "github.com/couchbaselabs/logg"

type InProcessJobRunner struct {
}

func NewInProcessJobRunner() *InProcessJobRunner {
	return &InProcessJobRunner{}
}

func (j InProcessJobRunner) ScheduleJob(job JobDescriptor) error {

	// create job locally and fire off go-routine
	logg.Log("in process job runner called with: %+v", job)

	return nil
}
