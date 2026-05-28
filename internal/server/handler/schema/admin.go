package schema

type GetInviteCodeReq struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type InviteCode struct {
	Code   string `json:"code"`
	Expire int64  `json:"expire"`
}

type UpdatePasswdReq struct {
	Old string `json:"old"`
	New string `json:"new"`
}
