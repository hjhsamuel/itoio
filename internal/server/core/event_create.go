package core

import (
	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/internal/server/room"
)

type EventCreate struct {
	Secret string `json:"secret"`
}

func (e *EventCreate) Execute(c *Core, info *ConnBase) error {
	var name string
	if userObj, err := c.d.GetUserByID(info.ID); err != nil {
		name = info.ID
	} else {
		name = userObj.Name
	}

	roomId, err := c.rooms.CreateRoom(e.Secret, &room.RoomUser{
		ID:       info.ID,
		Nickname: name,
		Owner:    true,
		Write:    info.Write,
	})
	if err != nil {
		Write(info.Write, &entities.EnterRoom{OK: false, Data: err.Error()})
		return err
	}
	Write(info.Write, &entities.EnterRoom{OK: true, Data: roomId})
	if roomObj, err := c.rooms.GetCurrentRoom(info.ID); err == nil {
		c.roomInfoChanged(roomObj)
	}
	return nil
}
