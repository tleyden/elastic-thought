package elasticthought

import "github.com/couchbaselabs/logg"

// The logging keys available in elastic thought
var LogKeys []string

func init() {
	LogKeys = []string{
		"CLI",
		"REST",
		"ELASTIC_THOUGHT",
		"MODEL",
		"CLASSIFIER",
		"CHANGES",
		"JOB_SCHEDULER",
		"NSQ_WORKER",
		"DATASET_SPLITTER",
		"DATAFILE_DOWNLOADER",
		"SOLVER",
		"TRAINING_JOB",
		"CLASSIFY_JOB",
		"TEST",
		"CBFS",
	}
}

// Enable logging for all logging keys
func EnableAllLogKeys() {
	for _, logKey := range LogKeys {
		logg.LogKeys[logKey] = true
	}

}
