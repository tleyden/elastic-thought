package elasticthought

import (
	"fmt"
	"html"
	"net/http"

	"github.com/dustin/go-couch"
	"github.com/gorilla/mux"
)

// A container to hold context associated with a REST API
// server.  This is used to avoid re-creating a handle to the
// database for each request, since it is an expensive operation.
type RestApiServer struct {
	DatabaseURL string         // the couchbase sync gw db url
	Database    couch.Database // the couchbase sync gw database handle

}

// Create a new REST API server and connect to the database.
func NewRestApiServer(dbUrl string) (*RestApiServer, error) {
	r := &RestApiServer{
		DatabaseURL: dbUrl,
	}
	err := r.connectDb()
	return r, err
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
		fmt.Fprintf(w, "Foo")
	}

	r.HandleFunc("/foo", fooHandler)
	r.HandleFunc("/", homeHandler)

	return r

}

func (s *RestApiServer) connectDb() error {
	db, err := couch.Connect(s.DatabaseURL)
	if err != nil {
		return err
	}
	s.Database = db
	return nil

}
