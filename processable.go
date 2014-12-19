package elasticthought

import (
	"github.com/couchbaselabs/logg"
	"github.com/dustin/httputil"
	"github.com/tleyden/go-couch"
)

type Processable interface {
	GetProcessingState() ProcessingState
	SetProcessingState(newState ProcessingState)
	RefreshFromDB(db couch.Database) error
}

func CasUpdateProcessingState(p Processable, newState ProcessingState, db couch.Database) (bool, error) {

	// if already has the newState, return false
	if p.GetProcessingState() == newState {
		logg.LogTo("TRAINING_JOB", "Already in state: %v", p.GetProcessingState())
		return false, nil
	}

	for {
		p.SetProcessingState(newState)

		// SAVE: try to save to the database
		logg.LogTo("TRAINING_JOB", "Trying to save: %+v", p)

		_, err := db.Edit(p)

		if err != nil {

			logg.LogTo("TRAINING_JOB", "Got error updating: %v", err)

			// if it failed with any other error than 409, return an error
			if !httputil.IsHTTPStatus(err, 409) {
				logg.LogTo("TRAINING_JOB", "Not a 409 error: %v", err)
				return false, err
			}

			// it failed with 409 error
			logg.LogTo("TRAINING_JOB", "Its a 409 error: %v", err)

			// get the latest version of the document
			if err := p.RefreshFromDB(db); err != nil {
				return false, err
			}

			logg.LogTo("TRAINING_JOB", "Retrieved new: %+v", p)

			// does it already have the new the state (eg, someone else set it)?
			if p.GetProcessingState() == newState {
				logg.LogTo("TRAINING_JOB", "Processing state already set")
				return false, nil
			}

			// no, so try updating state and saving again
			continue

		}

		// ensure that by the time we return, the processable has the most
		// version from the db
		if err := p.RefreshFromDB(db); err != nil {
			return false, err
		}

		// successfully saved, we are done
		logg.LogTo("TRAINING_JOB", "Successfully saved: %+v", p)
		return true, nil

	}

}
