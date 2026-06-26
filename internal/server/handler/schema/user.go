package schema

type LoginReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Device   string `json:"device"`   // device id
	DevName  string `json:"dev_name"` // device host name
}

type LoginRsp struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Expire   int    `json:"expire"`
}

type RefreshRsp struct {
	Expire int `json:"expire"`
}

type RegisterReq struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}
