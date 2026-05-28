package core

type EventLeave struct{}

func (e *EventLeave) Execute(c *Core, info *ConnBase) error {
	c.cancelDisconnectTimer(info.ID)
	return c.leaveRoom(info.ID, "user leave")
}
