package elasticthought

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/couchbaselabs/logg"
	"github.com/dustin/go-couch"
	"github.com/gorilla/mux"
)

// A container to hold settings associated with a REST API
// server.
type RestApiServer struct {
	DatabaseURL string // the couchbase sync gw db url

}

// a handler that is passed in a database handle in addition to req/res
type dbHandlerFunc func(http.ResponseWriter, *http.Request, couch.Database)

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
		fmt.Fprintf(w, "Welcome to the ElasticThought REST API")
	}

	r.HandleFunc("/users", s.createHandler(handleNewUser)).Methods("POST")
	r.HandleFunc("/datafiles", s.createHandler(handleNewDatafile)).Methods("POST")
	r.HandleFunc("/", homeHandler)

	return r

}

func handleNewDatafile(w http.ResponseWriter, r *http.Request, db couch.Database) {

	if loggedInUser, err := getBasicAuthUser(r); err != nil {
		errMsg := fmt.Sprintf("Unable to authenticate user: %v", err)
		http.Error(w, errMsg, 500)
		return
	}

	// create a new Datfile object

	// _changes listener will see it and process it (download and save to s3)

	// return uuid of Dataafile object

}

func handleNewUser(w http.ResponseWriter, r *http.Request, db couch.Database) {

	// parse in a user object from the POST request
	decoder := json.NewDecoder(r.Body)
	userToCreate := NewUser()
	err := decoder.Decode(userToCreate)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to parse user params: %v", err)
		http.Error(w, errMsg, 500)
		return
	}

	// make sure this user isn't already in the db
	existingUser := NewUser()
	err = db.Retrieve(userToCreate.DocId(), existingUser)
	if err == nil {
		errMsg := fmt.Sprintf("User already exists: %+v", *existingUser)
		http.Error(w, errMsg, 500)
		return
	}

	logg.LogTo("REST", "Did not find existing user, ok to create")

	// create a new user and return 201
	newUser := NewUserFromUser(*userToCreate)
	id, rev, err := db.InsertWith(newUser, newUser.DocId())
	if err != nil {
		errMsg := fmt.Sprintf("Error creating new user: %v", err)
		http.Error(w, errMsg, 500)
		return
	}

	fmt.Fprintf(w, "Created new user with id: %v rev: %v", id, rev)

}

// wrap a db handler func with a HandlerFunc
func (s RestApiServer) createHandler(dbHandler dbHandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// create a connection to the db -- for now, creates a new one
		// for handling each request.
		db, err := s.getDbConnection()
		if err != nil {
			errMsg := fmt.Sprintf("Unable to connect to DB: %v", err)
			http.Error(w, errMsg, 500)
			return
		}
		dbHandler(w, r, db)
	}

}

func (s RestApiServer) getDbConnection() (couch.Database, error) {
	db, err := couch.Connect(s.DatabaseURL)
	if err != nil {
		return db, err
	}
	return db, nil

}
