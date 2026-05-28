package schema

import "fmt"

const (
	UserPattern = "user:*"

	IdxUserName = "idx_user_name"
)

var UserKey = func(id string) string { return fmt.Sprintf("user:%s", id) }

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}
