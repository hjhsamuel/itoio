package schema

import "fmt"

const (
	InviteCodePattern = "invite:*"

	IdxInvitor = "idx_invite_user"
)

var InviteCodeKey = func(code string) string { return fmt.Sprintf("invite:%s", code) }

type InviteCode struct {
	Code   string `json:"code"`
	User   string `json:"user"`
	Expire int64  `json:"expire"`
}
