package elasticthought

// TODO: rename to JobScheduler
type JobRunner interface {
	ScheduleJob(job JobDescriptor) error
}
