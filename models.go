package elasticthought

const (
	DOC_TYPE_USER         = "user"
	DOC_TYPE_DATAFILE     = "datafile"
	DOC_TYPE_DATASET      = "dataset"
	DOC_TYPE_SOLVER       = "solver"
	DOC_TYPE_TRAINING_JOB = "training-job"
)

// All document structs should embed this struct go get access to
// the sync gateway metadata (_id, _rev) and the "type" field
// which differentiates the different doc types.
type ElasticThoughtDoc struct {
	Revision string `json:"_rev"`
	Id       string `json:"_id"`
	Type     string `json:"type"`
}
