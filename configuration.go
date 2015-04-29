package elasticthought

import (
	"errors"
	"fmt"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

type QueueType int

const (
	Nsq QueueType = iota
	Goroutine
)

// Holds configuration values that are used throughout the application
type Configuration struct {
	DbUrl               string
	CbfsUrl             string
	NsqLookupdUrl       string
	NsqdUrl             string
	NsqdTopic           string
	WorkDirectory       string
	QueueType           QueueType
	NumCbfsClusterNodes int // needed to validate cbfs cluster health
}

func NewDefaultConfiguration() *Configuration {

	config := &Configuration{
		DbUrl:               "http://localhost:4985/elastic-thought",
		CbfsUrl:             "file:///tmp",
		NsqLookupdUrl:       "127.0.0.1:4161",
		NsqdUrl:             "127.0.0.1:4150",
		NsqdTopic:           "elastic-thought",
		WorkDirectory:       "/tmp/elastic-thought",
		QueueType:           Goroutine,
		NumCbfsClusterNodes: 1,
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

// Create a new cbfs client based on url stored in config
func (c Configuration) NewBlobStoreClient() (BlobStore, error) {
	return NewBlobStore(c.CbfsUrl)
}

// Add values from parsedDocOpts into Configuration and return a new instance
// Example map:
//     map[--help:false --blob-store-url:file:///tmp --sync-gw-url:http://blah.com:4985/et]
func (c Configuration) Merge(parsedDocOpts map[string]interface{}) (Configuration, error) {

	// Sync Gateway URL
	syncGwUrl, ok := parsedDocOpts["--sync-gw-url"].(string)
	if !ok {
		return c, fmt.Errorf("Expected string arg in --sync-gw-url, got %T", syncGwUrl)
	}
	c.DbUrl = syncGwUrl

	// Blob Store URL
	blobStoreUrl, ok := parsedDocOpts["--blob-store-url"].(string)
	if !ok {
		return c, fmt.Errorf("Expected string arg in --blob-store-url, got %T", blobStoreUrl)
	}
	c.CbfsUrl = blobStoreUrl

	return c, nil

}
