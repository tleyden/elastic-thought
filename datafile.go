package elasticthought

import "github.com/tleyden/go-couch"

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

func FindDatafile(db couch.Database, datafileId string) (*Datafile, error) {

	datafile := &Datafile{}
	if err := db.Retrieve(datafileId, datafile); err != nil {
		return nil, err
	}
	return datafile, nil

}
