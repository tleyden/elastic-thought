package elasticthought

import "github.com/tleyden/go-couch"

type Processable interface {
	GetProcessingState() ProcessingState
	SetProcessingState(newState ProcessingState)
	RefreshFromDB(db couch.Database) error
}
