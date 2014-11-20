package elasticthought

import (
	"fmt"

	"github.com/tleyden/go-couch"
)

// A solver can generate trained models, which ban be used to make predictions
type Solver struct {
	ElasticThoughtDoc
	DatasetId        string            `json:"dataset-id"`
	SpecificationUrl string            `json:"specification-url" binding:"required"`
	SpecificationEnv map[string]string `json:"specification-env"`
}

// Create a new solver.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewSolver() *Solver {
	return &Solver{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_SOLVER},
	}
}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
func (s Solver) Insert(db couch.Database) (*Solver, error) {

	id, _, err := db.Insert(s)
	if err != nil {
		err := fmt.Errorf("Error inserting solver: %v.  Err: %v", s, err)
		return nil, err
	}

	// load dataset object from db (so we have id/rev fields)
	solver := &Solver{}
	err = db.Retrieve(id, solver)
	if err != nil {
		err := fmt.Errorf("Error fetching solver: %v.  Err: %v", id, err)
		return nil, err
	}

	return solver, nil

}
