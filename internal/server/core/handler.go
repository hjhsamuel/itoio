package core

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/internal/server/handler/request"
	"github.com/hjhsamuel/itoio/internal/server/room"
	"github.com/sirupsen/logrus"
)

const (
	pongWait = 20 * time.Second
)

func (c *Core) WebSocketHandler(user *request.User, conn *websocket.Conn) {
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	client := NewWebSocketConn(user.ID, user.Device, conn, c.read)
	c.read <- &ConnMessage{
		Info:  client.info,
		Event: &EventConnect{},
	}
	client.Start()
}

func (c *Core) roomInfoChanged(r *room.Room) {
	users := make([]*entities.User, 0, len(r.Users))
	for _, user := range r.Users {
		users = append(users, &entities.User{
			ID:       user.ID,
			Nickname: user.Nickname,
			Owner:    user.Owner,
		})
	}
	for _, user := range r.Users {
		if user.Write == nil {
			continue
		}
		Write(user.Write, &entities.Room{ID: r.ID, Users: users})
	}
}

func (c *Core) leaveRoom(userID, reason string) error {
	roomObj, err := c.rooms.GetCurrentRoom(userID)
	if err != nil {
		return nil
	}
	closedRoom := c.rooms.LeaveRoom(userID)
	if closedRoom {
		for _, user := range roomObj.Users {
			if user.Write != nil {
				Write(user.Write, &entities.RoomClosed{Room: roomObj.ID, Reason: reason})
			}
		}
		c.rooms.CloseRoom(roomObj.ID)
		return nil
	}
	if len(roomObj.Users) == 0 {
		c.rooms.CloseRoom(roomObj.ID)
		return nil
	}
	c.roomInfoChanged(roomObj)
	return nil
}

func (c *Core) do() {
	defer c.wg.Done()
	for {
		select {
		case msg := <-c.read:
			err := msg.Event.Execute(c, msg.Info)
			if err != nil {
				logrus.Error(err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}
