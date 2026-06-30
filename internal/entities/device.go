package entities

type DeviceInfo struct {
	ID    string `json:"id"`
	User  string `json:"user"`
	Name  string `json:"name"`
	State int    `json:"state"`
	Temp  bool   `json:"temp"`
	Room  string `json:"room"`
}
