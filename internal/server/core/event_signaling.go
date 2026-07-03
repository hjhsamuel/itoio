package core

import (
	"fmt"
	"strings"

	"github.com/hjhsamuel/itoio/internal/entities"
)

type EventSignaling struct {
	From string `json:"from"`
	To   string `json:"to"`
	Typ  string `json:"typ"`
	Data any    `json:"data"`
}

func (e *EventSignaling) Execute(c *Core, info *ConnBase) error {
	fmt.Println(e.From, e.To, e.Typ)
	var fromID string
	if !strings.Contains(e.From, ":") {
		// browser
		fromID = info.UserID
	} else {
		fromID = info.UUID
	}
	var toID string
	if strings.Contains(e.To, ":") {
		toID = e.To
	} else {
		toID = e.To + ":"
	}

	if e.From != fromID {
		return fmt.Errorf("signaling sender %s does not match current connection %s", e.From, fromID)
	}
	if !c.rooms.InSameRoom(info.UUID, toID) {
		return fmt.Errorf("users %s and %s are not in the same room", fromID, toID)
	}
	conn := c.client[toID]
	if conn == nil {
		return fmt.Errorf("target user %s is not connected", toID)
	}
	Write(conn.Write, &entities.Signaling{
		Typ:  e.Typ,
		From: e.From,
		To:   e.To,
		Data: e.Data,
	})
	return nil
}
