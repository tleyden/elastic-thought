package elasticthought

import (
	"encoding/json"
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestJsonDecode(t *testing.T) {

	jsonString := `
	    {  
	       "_id":"id",
	       "_rev":"rev",
	       "type":"dataset",
	       "datafile-id":"123",
	       "training":{  
		  "split-percentage":0.7
	       },
	       "test":{  
		  "split-percentage":0.3
	       },
	       "processing-state":"processing"
	    }`

	data := []byte(jsonString)

	dataset := Dataset{}
	err := json.Unmarshal(data, &dataset)

	assert.True(t, err == nil)

	assert.Equals(t, dataset.ProcessingState, Processing)

}

func TestJsonEncode(t *testing.T) {

	configuration := NewDefaultConfiguration()

	// create a dataset struct
	dataset := NewDataset(*configuration)
	dataset.ProcessingState = FinishedSuccessfully
	dataset.TrainingDataset.SplitPercentage = 0.7
	dataset.TrainingDataset.DatafileID = "dfid"
	dataset.TestDataset.SplitPercentage = 0.3
	dataset.TestDataset.DatafileID = "dfid"
	dataset.Id = "dsid"

	// marshal dataset -> json
	data, err := json.Marshal(dataset)
	assert.True(t, err == nil)

	// now try to parse the json back into a struct
	dataset2 := Dataset{}
	err = json.Unmarshal(data, &dataset2)
	assert.True(t, err == nil)

	// make assertions
	assert.Equals(t, dataset2.ProcessingState, FinishedSuccessfully)

}
