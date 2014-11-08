package elasticthought

import (
	"errors"
	"fmt"

	"github.com/tleyden/go-couch"
)

const (
	DOC_TYPE_USER     = "user"
	DOC_TYPE_DATAFILE = "datafile"
)

type ElasticThoughtDoc struct {
	Revision string `json:"_rev"`
	Id       string `json:"_id"`
	Type     string `json:"type"`
}

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

func (u User) DocId() string {
	return docIdFromUsername(u.Username)
}

func docIdFromUsername(username string) string {
	return fmt.Sprintf("user:%v", username)
}

// A Datafile
type Datafile struct {
	ElasticThoughtDoc
	UserID string `json:"user-id"`
	Url    string `json:"url" binding:"required"`
}
