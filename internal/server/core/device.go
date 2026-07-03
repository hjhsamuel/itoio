package core

import (
	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/hjhsamuel/itoio/internal/entities"
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

func (c *Core) GetDevices(userID string, page, limit int) ([]*entities.DeviceInfo, int64, error) {
	objs, cnt, err := c.d.GetDevicesByUser(userID)
	if err != nil {
		return nil, 0, err
	}
	var (
		minIndex = (page - 1) * limit
		maxIndex = page * limit
	)
	res := make([]*entities.DeviceInfo, 0)
	for i, item := range objs {
		if i < minIndex {
			continue
		}
		if i >= maxIndex {
			break
		}
		obj := &entities.DeviceInfo{
			ID:    item.Device,
			User:  item.User,
			Name:  item.Name,
			State: int(item.State),
			Temp:  item.Temp,
		}
		if item.State == schema.DeviceStateOnline {
			id := GenerateConnUUID(item.User, item.Device)
			room, err := c.rooms.GetCurrentRoom(id)
			if err == nil {
				obj.Room = room.ID
			} else {
				obj.State = int(schema.DeviceStateOffline)
			}
		}
		res = append(res, obj)
	}
	return res, cnt, nil
}

func (c *Core) AddDeviceIfNotExist(userID, deviceId string, name string) error {
	_, err := c.d.GetDevice(userID, deviceId)
	if err == nil {
		return nil
	}
	return c.AddDevice(userID, deviceId, name)
}
