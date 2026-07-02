package schema

import "fmt"

const (
	TokenPattern = "token:*"

	IdxTokenData = "idx_token_data"
)

var TokenKey = func(userId, deviceId string) string { return fmt.Sprintf("token:%s:%s", userId, deviceId) }

type Token struct {
	User   string `json:"user"`
	Device string `json:"device"`
	Data   string `json:"data"`
}
