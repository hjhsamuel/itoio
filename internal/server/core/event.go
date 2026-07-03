package core

import "encoding/json"

const (
	EventTypeOffer     = "offer"     // peer send offer
	EventTypeAnswer    = "answer"    // peer send answer
	EventTypeCandidate = "candidate" // peer send candidate

	EventTypeCreate = "create" // user create room
	EventTypeJoin   = "join"   // user join room
	EventTypeLeave  = "leave"  // user leave room

	EventTypeStopShare = "stop_share" // user stop share

	EventTypeCJoin = "cjoin" // user join control room
)

type Request struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Event interface {
	Execute(c *Core, info *ConnBase) error
}
