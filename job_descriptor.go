package elasticthought

// A job descriptor is meant to fully describe a worker job.  It needs to be
// easily serializable into json so it can be passed on the wire.
type JobDescriptor struct {
	DocIdToProcess string `json:"doc-id-to-process"`
}

// Create a new JobDescriptor
func NewJobDescriptor(docId string) *JobDescriptor {
	return &JobDescriptor{
		DocIdToProcess: docId,
	}
}
