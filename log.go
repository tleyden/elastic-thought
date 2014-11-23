package elasticthought

import "github.com/couchbaselabs/logg"

// The logging keys available in elastic thought
var LogKeys []string

func init() {
	LogKeys = []string{
		"CLI",
		"REST",
		"CHANGES",
		"JOB_SCHEDULER",
		"NSQ_WORKER",
		"DATASET_SPLITTER",
		"SOLVER",
		"TRAINING_JOB",
		"TEST",
	}
}

// Enable logging for all logging keys
func EnableAllLogKeys() {
	for _, logKey := range LogKeys {
		logg.LogKeys[logKey] = true
	}

}
