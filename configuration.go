package elasticthought

import (
	"errors"
	"fmt"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// Holds configuration values that are used throughout the application
type Configuration struct {
	DbUrl         string
	NsqLookupdUrl string
	NsqdUrl       string
	NsqdTopic     string
}

// Connect to db based on url stored in config, or panic if not able to connect
func (c Configuration) DbConnection() couch.Database {
	db, err := couch.Connect(c.DbUrl)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error %v | dbUrl: %v", err, c.DbUrl))
		logg.LogPanic("%v", err)
	}
	return db
}
