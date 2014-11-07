package elasticthought

import "fmt"

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

func (u User) DocId() string {
	return fmt.Sprintf("user:%v", u.Username)
}
