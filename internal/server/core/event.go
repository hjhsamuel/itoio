package core

import "encoding/json"

const (
	EventTypeOffer     = "offer"
	EventTypeAnswer    = "answer"
	EventTypeCandidate = "candidate"

	EventTypeCreate = "create"
	EventTypeJoin   = "join"
	EventTypeLeave  = "leave"

	EventTypeStopShare = "stop_share"

	EventTypeCJoin = "cjoin"
)

type Request struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Event interface {
	Execute(c *Core, info *ConnBase) error
}
