package elasticthought

import "fmt"

// Create a new job based on the Job Descriptor
func CreateJob(config Configuration, jobDescriptor JobDescriptor) (Runnable, error) {

	// Connect to DB
	db := config.DbConnection()

	// Fetch doc associated w/ job descriptor
	doc := &ElasticThoughtDoc{}
	err := db.Retrieve(jobDescriptor.DocIdToProcess, doc)
	if err != nil {
		return nil, err
	}

	// based on document type, create the correct Runnable
	switch doc.Type {
	case DOC_TYPE_DATASET:
		return DatasetSplitter{}, nil
	}

	return nil, fmt.Errorf("Unable to create job for: %+v", jobDescriptor)

}
