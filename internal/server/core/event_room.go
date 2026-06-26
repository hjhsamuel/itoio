package core

import "errors"

type EventRoom struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func (e *EventRoom) Execute(c *Core, info *ConnBase) error {
	if e.Secret == "" {
		return errors.New("secret is empty")
	}
	return c.rooms.UpdateSecret(e.ID, info.UUID, e.Secret)
}
