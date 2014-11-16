package elasticthought

// Holds configuration values that are used throughout the application
type Configuration struct {
	DbUrl         string
	NsqLookupdUrl string
	NsqdUrl       string
	NsqdTopic     string
}
