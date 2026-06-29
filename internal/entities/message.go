package entities

import "encoding/json"

const (
	MessageTypeCloseConn = "close"

	MessageTypeRoom       = "room"
	MessageTypeEnter      = "enter"
	MessageTypeRoomClosed = "room_closed"
	MessageTypeIceConfig  = "ice_config"

	MessageTypeOffer     = "offer"
	MessageTypeAnswer    = "answer"
	MessageTypeCandidate = "candidate"

	MessageTypeControl = "control"
)

type Message interface {
	Type() string
}

type P2PMessage struct {
	ID   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

type CloseConn struct {
	Code int
	Msg  string
}

func (c *CloseConn) Type() string {
	return MessageTypeCloseConn
}

type Room struct {
	ID    string  `json:"id"`
	Users []*User `json:"users"`
}

type User struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Owner    bool   `json:"owner"`
}

type IceServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type IceConfig struct {
	IceServers []*IceServer `json:"ice_servers"`
	Mode       string       `json:"mode"`
}

func (c *IceConfig) Type() string {
	return MessageTypeIceConfig
}

func (r *Room) Type() string {
	return MessageTypeRoom
}

type EnterRoom struct {
	OK   bool   `json:"ok"`
	Data string `json:"data"` // room id or error message
}

func (r *EnterRoom) Type() string {
	return MessageTypeEnter
}

type RoomClosed struct {
	Room   string `json:"room"`
	Reason string `json:"reason"`
}

func (r *RoomClosed) Type() string {
	return MessageTypeRoomClosed
}

type Signaling struct {
	Typ  string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
	Data any    `json:"data"`
}

func (s *Signaling) Type() string {
	return s.Typ
}

type Control struct{}

func (c *Control) Type() string {
	return MessageTypeControl
}
