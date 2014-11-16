package elasticthought

// A job descriptor is meant to fully describe a worker job.  It needs to be
// easily serializable into json so it can be passed on the wire.
type JobDescriptor struct {
	DocIdToProcess string        `json:"doc-id-to-process"`
	Configuration  Configuration `json:"configuration"`
}

// Create a new JobDescriptor
func NewJobDescriptor(c Configuration, docId string) *JobDescriptor {
	return &JobDescriptor{
		Configuration:  c,
		DocIdToProcess: docId,
	}
}
