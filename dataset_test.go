package elasticthought

import (
	"encoding/json"
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
)

func TestJsonDecode(t *testing.T) {

	jsonString := `{"_id":"id","_rev":"rev","type":"dataset","datafile-id":"123","split-percentage":0.7,"processing-state":"processing"}`

	data := []byte(jsonString)

	dataset := Dataset{}
	err := json.Unmarshal(data, &dataset)

	assert.True(t, err == nil)

	assert.Equals(t, dataset.ProcessingState, Processing)

}

func TestJsonEncode(t *testing.T) {

	// create a dataset struct
	dataset := NewDataset()
	dataset.DatafileID = "dfid"
	dataset.ProcessingState = FinishedSuccessfully
	dataset.SplitPercentage = 0.7
	dataset.Id = "dsid"

	// marshal dataset -> json
	data, err := json.Marshal(dataset)
	assert.True(t, err == nil)

	// now try to parse the json back into a struct
	dataset2 := Dataset{}
	err = json.Unmarshal(data, &dataset2)
	logg.Log("Err: %v", err)
	assert.True(t, err == nil)

	// make assertions
	assert.Equals(t, dataset2.ProcessingState, FinishedSuccessfully)

}
