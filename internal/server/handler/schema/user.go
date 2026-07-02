package schema

type LoginReq struct {
	Name     string       `json:"name"`
	Password string       `json:"password"`
	Device   *LoginDevice `json:"device,omitempty"`
}

type LoginDevice struct {
	ID       string `json:"id"`       // device id
	Name     string `json:"name"`     // device host name
	Platform string `json:"platform"` // device platform
	OS       string `json:"os"`       // device os
	Ver      string `json:"ver"`      // device version
}

type LoginRsp struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Expire   int    `json:"expire"`
	Token    string `json:"token"`
}

type RefreshReq struct {
	Token string `json:"token,omitempty"`
}

type RefreshRsp struct {
	Expire int    `json:"expire"`
	Token  string `json:"token"`
}

type RegisterReq struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}
