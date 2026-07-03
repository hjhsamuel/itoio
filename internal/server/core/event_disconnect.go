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

type EventDisconnectExpired struct{}

func (e *EventDisconnect) Execute(c *Core, info *ConnBase) error {
	current, ok := c.client[info.UUID]
	if !ok || current != info {
		return nil
	}
	delete(c.client, info.UUID)
	Write(info.Write, &entities.CloseConn{Code: e.Code, Msg: e.Msg})

	if info.Device != "" {
		_ = c.leaveRoom(info.UUID, "device disconnect")
	} else {
		if _, err := c.rooms.MarkDisconnected(info.UUID); err != nil {
			return nil
		}
		c.scheduleDisconnectExpiry(info.UserID, info.Device, info.UUID)
	}

	return nil
}

func (e *EventDisconnectExpired) Execute(c *Core, info *ConnBase) error {
	if _, ok := c.client[info.UUID]; ok {
		return nil
	}
	delete(c.disconnectTimers, info.UUID)
	if info.Device != "" {
		_ = c.d.DeviceOffline(info.UserID, info.Device)
	}
	return c.leaveRoom(info.UUID, "reconnect timeout")
}

func (c *Core) scheduleDisconnectExpiry(userId, device, uid string) {
	c.cancelDisconnectTimer(uid)
	c.disconnectTimers[uid] = time.AfterFunc(reconnectGracePeriod, func() {
		c.read <- &ConnMessage{
			Info:  &ConnBase{UserID: userId, Device: device, UUID: uid},
			Event: &EventDisconnectExpired{},
		}
	})
}

func (c *Core) cancelDisconnectTimer(id string) {
	timer, ok := c.disconnectTimers[id]
	if !ok {
		return
	}
	timer.Stop()
	delete(c.disconnectTimers, id)
}
