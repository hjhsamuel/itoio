package core

import (
	"errors"
	"strings"

	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/internal/server/room"
)

type EventCJoin struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func (e *EventCJoin) Execute(c *Core, info *ConnBase) error {
	if !strings.HasPrefix(e.ID, room.ControlPrefix) {
		return errors.New("invalid room id: " + e.ID)
	}

	var name string
	if userObj, err := c.d.GetUserByID(info.UserID); err != nil {
		name = info.UUID
	} else {
		name = userObj.Name
	}
	user := &room.RoomUser{
		ID:       info.UUID,
		Nickname: name,
		Owner:    false,
		Write:    info.Write,
	}
	roomObj, err := c.rooms.JoinRoom(e.ID, e.Secret, user)
	if err != nil {
		Write(info.Write, &entities.EnterRoom{OK: false, Data: err.Error()})
	} else {
		c.roomInfoChanged(roomObj)
	}
	return err
}
