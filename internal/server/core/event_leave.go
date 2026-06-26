package core

type EventLeave struct{}

func (e *EventLeave) Execute(c *Core, info *ConnBase) error {
	c.cancelDisconnectTimer(info.UUID)
	return c.leaveRoom(info.UUID, "user leave")
}
