package elasticthought

type JobRunner interface {
	ScheduleJob(job JobDescriptor) error
}
