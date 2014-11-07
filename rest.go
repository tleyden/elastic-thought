package elasticthought

import (
	"fmt"
	"html"
	"net/http"

	"github.com/dustin/go-couch"
	"github.com/gorilla/mux"
)

// A container to hold settings associated with a REST API
// server.
type RestApiServer struct {
	DatabaseURL string // the couchbase sync gw db url

}

// Create a new REST API server and connect to the database.
func NewRestApiServer(dbUrl string) *RestApiServer {
	r := &RestApiServer{
		DatabaseURL: dbUrl,
	}
	return r
}

// Get the ElasticThought handler.  This is de-coupled from the
// webserver startup in case you want to embed ElasticThought into
// another webserver.
func (s RestApiServer) RestApiRouter() *mux.Router {

	r := mux.NewRouter()

	homeHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	}
	fooHandler := func(w http.ResponseWriter, r *http.Request) {
		db, err := s.getDbConnection()
		fmt.Fprintf(w, "DB: %v Error: %v", db, err)
	}

	r.HandleFunc("/foo", fooHandler)
	r.HandleFunc("/", homeHandler)

	return r

}

func (s RestApiServer) getDbConnection() (couch.Database, error) {
	db, err := couch.Connect(s.DatabaseURL)
	if err != nil {
		return db, err
	}
	return db, nil

}
