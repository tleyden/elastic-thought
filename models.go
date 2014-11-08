package elasticthought

import (
	"errors"
	"fmt"

	"github.com/dustin/go-couch"
)

// An ElasticThought user.
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create a new User
func NewUser() *User {
	return &User{}
}

// Create a new User based on values in another user
func NewUserFromUser(other User) *User {
	user := &User{}
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
