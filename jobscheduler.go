package elasticthought

// By swapping out the job scheduler, you can easily swap out the ability
// to have jobs placed on NSQ or to be run in a local goroutine inside the
// rest server process
type JobScheduler interface {
	ScheduleJob(job JobDescriptor) error
}
