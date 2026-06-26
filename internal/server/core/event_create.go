package core

import (
	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/internal/server/room"
)

type EventCreate struct {
	Secret string        `json:"secret"`
	Mode   room.RoomMode `json:"mode"`
}

func (e *EventCreate) Execute(c *Core, info *ConnBase) error {
	var name string
	if userObj, err := c.d.GetUserByID(info.UserID); err != nil {
		name = info.UUID
	} else {
		name = userObj.Name
	}

	roomId, err := c.rooms.CreateRoom(e.Secret, e.Mode, &room.RoomUser{
		ID:       info.UUID,
		Nickname: name,
		Owner:    true,
		Write:    info.Write,
	})
	if err != nil {
		Write(info.Write, &entities.EnterRoom{OK: false, Data: err.Error()})
		return err
	}
	Write(info.Write, &entities.EnterRoom{OK: true, Data: roomId})
	if roomObj, err := c.rooms.GetCurrentRoom(info.UUID); err == nil {
		c.roomInfoChanged(roomObj)
	}
	return nil
}
