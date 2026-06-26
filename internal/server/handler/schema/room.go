package schema

type UpdateRoomConfigReq struct {
	Room   string `json:"room"`
	Secret string `json:"secret"`
}
