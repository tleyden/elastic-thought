package elasticthought

// A Datafile
type Datafile struct {
	ElasticThoughtDoc
	UserID string `json:"user-id"`
	Url    string `json:"url" binding:"required"`
}
