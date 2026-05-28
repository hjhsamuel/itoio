package core

import (
	"fmt"

	"github.com/hjhsamuel/itoio/internal/entities"
)

type EventSignaling struct {
	From string `json:"from"`
	To   string `json:"to"`
	Typ  string `json:"typ"`
	Data any    `json:"data"`
}

func (e *EventSignaling) Execute(c *Core, info *ConnBase) error {
	if e.From != info.ID {
		return fmt.Errorf("signaling sender %s does not match current connection %s", e.From, info.ID)
	}
	if !c.rooms.InSameRoom(e.From, e.To) {
		return fmt.Errorf("users %s and %s are not in the same room", e.From, e.To)
	}
	conn := c.client[e.To]
	if conn == nil {
		return fmt.Errorf("target user %s is not connected", e.To)
	}
	Write(conn.Write, &entities.Signaling{
		Typ:  e.Typ,
		From: e.From,
		To:   e.To,
		Data: e.Data,
	})
	return nil
}
