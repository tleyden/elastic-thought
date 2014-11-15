package elasticthought

// A Datafile is a raw "bundle" of data, typically a zip or .tar.gz file.
// It cannot be used by a solver directly, instead it used to create
// dataset objects which can be used by the solver.
// A single datafile can be used to create any number of dataset objects.
type Datafile struct {
	ElasticThoughtDoc
	UserID string `json:"user-id"`
	Url    string `json:"url" binding:"required"`
}

// Create a new datafile
func NewDatafile() *Datafile {
	return &Datafile{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_DATAFILE},
	}
}
