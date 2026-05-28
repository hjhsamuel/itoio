package core

import (
	"time"

	"github.com/hjhsamuel/itoio/internal/entities"
)

const reconnectGracePeriod = 15 * time.Second

type EventDisconnect struct {
	Code int
	Msg  string
}

type EventDisconnectExpired struct {
	ID string
}

func (e *EventDisconnect) Execute(c *Core, info *ConnBase) error {
	current, ok := c.client[info.ID]
	if !ok || current != info {
		return nil
	}
	delete(c.client, info.ID)
	Write(info.Write, &entities.CloseConn{Code: e.Code, Msg: e.Msg})

	if _, err := c.rooms.MarkDisconnected(info.ID); err != nil {
		return nil
	}
	c.scheduleDisconnectExpiry(info.ID)
	return nil
}

func (e *EventDisconnectExpired) Execute(c *Core, info *ConnBase) error {
	if _, ok := c.client[e.ID]; ok {
		return nil
	}
	delete(c.disconnectTimers, e.ID)
	return c.leaveRoom(e.ID, "reconnect timeout")
}

func (c *Core) scheduleDisconnectExpiry(userID string) {
	c.cancelDisconnectTimer(userID)
	c.disconnectTimers[userID] = time.AfterFunc(reconnectGracePeriod, func() {
		c.read <- &ConnMessage{
			Info:  &ConnBase{ID: userID},
			Event: &EventDisconnectExpired{ID: userID},
		}
	})
}

func (c *Core) cancelDisconnectTimer(userID string) {
	timer, ok := c.disconnectTimers[userID]
	if !ok {
		return
	}
	timer.Stop()
	delete(c.disconnectTimers, userID)
}
