package elasticthought

import (
	"sync"

	"github.com/couchbaselabs/logg"
)

// Run worker jobs in a goroutine in the rest server process (as oppposed to using nsq)
// Makes certain testing easier
type InProcessJobScheduler struct {
	Configuration   Configuration
	JobsOutstanding *sync.WaitGroup
}

func NewInProcessJobScheduler(c Configuration) *InProcessJobScheduler {

	return &InProcessJobScheduler{
		Configuration:   c,
		JobsOutstanding: &sync.WaitGroup{},
	}
}

func (j InProcessJobScheduler) ScheduleJob(jobDescriptor JobDescriptor) error {

	// create job locally and fire off go-routine
	logg.LogTo("ELASTIC_THOUGHT", "in process job runner called with: %+v", jobDescriptor)

	// wait until there aren't any outstanding jobs (we want to do this
	// so that if we are processing a job already, we don't pick up new
	// jobs and give other peer httpd's the opportunity to pick up jobs)
	j.JobsOutstanding.Wait()

	job, err := CreateJob(j.Configuration, jobDescriptor)
	if err != nil {
		return err
	}

	j.JobsOutstanding.Add(1)

	go job.Run(j.JobsOutstanding)

	return nil
}
