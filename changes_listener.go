package elasticthought

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// A changes listener listens for changes on the _changes feed and reacts to them
type ChangesListener struct {
	DbUrl    string
	Database couch.Database
}

// Create a new ChangesListener
func NewChangesListener(dbUrl string) (*ChangesListener, error) {

	db, err := couch.Connect(dbUrl)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error %v | dbUrl: %v", err, dbUrl))
		logg.LogError(err)
		return nil, err
	}

	return &ChangesListener{
		DbUrl:    dbUrl,
		Database: db,
	}, nil
}

// Follow changes feed.  This will typically be run in its own goroutine.
func (c ChangesListener) FollowChangesFeed() {

	logg.LogTo("CHANGES", "going to follow changes feed")

	var since interface{}

	handleChange := func(reader io.Reader) interface{} {
		logg.LogTo("CHANGES", "handleChange() callback called")
		changes, err := decodeChanges(reader)
		if err != nil {
			// it's very common for this to timeout while waiting for new changes.
			// since we want to follow the changes feed forever, just log an error
			// TODO: don't even log an error if its an io.Timeout, just noise
			logg.LogTo("CHANGES", "%T decoding changes: %v.", err, err)
			return since
		}

		since = changes.LastSequence
		logg.LogTo("CHANGES", "returning since: %v", since)
		return since

	}

	options := map[string]interface{}{}
	options["feed"] = "longpoll"

	logg.LogTo("CHANGES", "Following changes feed: %+v", options)

	// this will block until the handleChange callback returns nil
	c.Database.Changes(handleChange, options)

	logg.LogPanic("Changes listener died -- this should never happen")

}

func decodeChanges(reader io.Reader) (couch.Changes, error) {

	changes := couch.Changes{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&changes)
	if err != nil {
		logg.LogTo("CHANGES", "Err decoding changes: %v", err)
	}
	return changes, err

}
