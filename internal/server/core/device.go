package core

import (
	"github.com/hjhsamuel/itoio/internal/dao/schema"
)

func (c *Core) AddDevice(userID, deviceId string, name string) error {
	device := &schema.Device{
		Device: deviceId,
		User:   userID,
		Name:   name,
		State:  schema.DeviceStateOnline,
		Temp:   true,
	}
	err := c.d.CreateDevice(device)
	if err != nil {
		return err
	}
	return nil
}

func (c *Core) UpdateDevice(userID, deviceId string, name *string, state schema.DeviceState) error {
	return c.d.UpdateDevice(userID, deviceId, name, state)
}

func (c *Core) DeleteDevice(id string, userID string) error {
	return c.d.DeleteDevice(id, userID)
}

func (c *Core) GetDevices(userID string, page, limit int) ([]*schema.Device, int64, error) {
	return c.d.GetDevicesByUser(userID, page, limit)
}
