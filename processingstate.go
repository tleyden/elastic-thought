package elasticthought

import (
	"encoding/json"
	"fmt"
)

// For objects that require processing, like Dataset objects, the ProcessingState
// helps track the current state of processing.
type ProcessingState int

const (
	Pending ProcessingState = iota
	Processing
	FinishedSuccessfully
	Failed
)

const (
	PROCESSING_STATE_PENDING               = "pending"
	PROCESSING_STATE_PROCESSING            = "processing"
	PROCESSING_STATE_FINISHED_SUCCESSFULLY = "finished_successfully"
	PROCESSING_STATE_FAILED                = "failed"
)

// Custom Unmarshal so that "Pending" is mapped to the numeric ProcessingState
func (p *ProcessingState) UnmarshalJSON(bytes []byte) error {

	var stringVal string
	err := json.Unmarshal(bytes, &stringVal)
	if err != nil {
		return err
	}

	switch stringVal {
	case PROCESSING_STATE_PENDING:
		*p = Pending
	case PROCESSING_STATE_PROCESSING:
		*p = Processing
	case PROCESSING_STATE_FINISHED_SUCCESSFULLY:
		*p = FinishedSuccessfully
	case PROCESSING_STATE_FAILED:
		*p = Failed
	default:
		return fmt.Errorf("Unexpected value for processing state: %v", stringVal)
	}

	return nil

}

func (p ProcessingState) MarshalJSON() ([]byte, error) {

	var stringVal string

	switch p {
	case Pending:
		stringVal = PROCESSING_STATE_PENDING
	case Processing:
		stringVal = PROCESSING_STATE_PROCESSING
	case FinishedSuccessfully:
		stringVal = PROCESSING_STATE_FINISHED_SUCCESSFULLY
	case Failed:
		stringVal = PROCESSING_STATE_FAILED
	default:
		return nil, fmt.Errorf("Unexpected value for processing state: %v", p)
	}

	// gotcha: string must be double quoted.  is there a cleaner approach?
	return []byte(fmt.Sprintf("\"%v\"", stringVal)), nil

}
