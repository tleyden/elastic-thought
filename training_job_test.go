package elasticthought

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
)

func TestRunCaffe(t *testing.T) {

	config := *NewDefaultConfiguration()
	trainingJob := NewTrainingJob()
	trainingJob.createWorkDirectory()
	trainingJob.Configuration = config

	trainingJob.Id = "training_job"
	err := trainingJob.runCaffe()
	logg.LogTo("TEST", "err: %v", err)
	assert.True(t, err == nil)

}
