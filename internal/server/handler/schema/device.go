package schema

type CreateDeviceReq struct {
	ID       string `json:"id"`       // device id
	Name     string `json:"name"`     // device host name
	Platform string `json:"platform"` // device platform
	OS       string `json:"os"`       // device os
	Ver      string `json:"ver"`      // device version
}

type UpdateDeviceReq struct {
	Name *string `json:"name"` // Device host name
}

type DelDeviceReq struct {
	ID string `json:"id"` // Device id
}

type MineDevicesReq struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type MineDeviceRsp struct {
	Total int64         `json:"total"`
	Data  []*DeviceInfo `json:"data"`
}

type DeviceInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State int    `json:"state"`
	Temp  bool   `json:"temp"`
	Room  string `json:"room"`
}
