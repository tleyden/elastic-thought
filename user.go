package elasticthought

import (
	"errors"
	"fmt"

	"github.com/tleyden/go-couch"
)

// An ElasticThought user.
type User struct {
	ElasticThoughtDoc
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create a new User
func NewUser() *User {
	return &User{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_USER},
	}
}

// Create a new User based on values in another user
func NewUserFromUser(other User) *User {
	user := NewUser()
	user.Username = other.Username
	user.Email = other.Email
	user.Password = other.Password
	return user
}

// Does this username/password combo exist in the database?  If so, return the
// user.  If not, return an error.
func AuthenticateUser(db couch.Database, username, password string) (*User, error) {

	userId := docIdFromUsername(username)

	authenticatedUser := NewUser()

	err := db.Retrieve(userId, authenticatedUser)

	if err != nil {
		// no valid user
		return nil, err
	}

	// we found a valid user, but do passwords match?
	if authenticatedUser.Password != password {
		err := errors.New("Passwords do not match")
		return nil, err
	}

	return authenticatedUser, err

}

// The doc id of this user.  If the username is "foo", the doc id will be "user:foo"
func (u User) DocId() string {
	return docIdFromUsername(u.Username)
}

func docIdFromUsername(username string) string {
	return fmt.Sprintf("user:%v", username)
}
