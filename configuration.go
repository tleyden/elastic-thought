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
	CbfsUrl       string
	NsqLookupdUrl string
	NsqdUrl       string
	NsqdTopic     string
}

func NewDefaultConfiguration() *Configuration {

	config := &Configuration{
		DbUrl:         "http://localhost:4985/elasticthought",
		CbfsUrl:       "http://localhost:8484",
		NsqLookupdUrl: "127.0.0.1:4161",
		NsqdUrl:       "127.0.0.1:4150",
		NsqdTopic:     "elastic-thought",
	}
	return config

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
