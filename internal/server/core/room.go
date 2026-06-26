package core

import "errors"

func (c *Core) UpdateRoomConfig(room, userId, deviceId string, secret string) error {
	if secret == "" {
		return errors.New("secret cannot be empty")
	}
	id := GenerateConnUUID(userId, deviceId)
	return c.rooms.UpdateSecret(room, id, secret)
}
