package dao

import (
	"encoding/json"
	"fmt"

	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/tidwall/buntdb"
)

func (d *Dao) CreateDevice(obj *schema.Device) error {
	key := schema.DeviceKey(obj.User, obj.Device)
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return d.d.Update(func(tx *buntdb.Tx) error {
		_, _, err = tx.Set(key, string(content), nil)
		return err
	})
}

func (d *Dao) DeviceOffline(userId, deviceId string) error {
	key := schema.DeviceKey(userId, deviceId)
	return d.d.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		var obj *schema.Device
		if err = json.Unmarshal([]byte(val), &obj); err != nil {
			return err
		}

		if obj.Temp {
			_, err = tx.Delete(key)
			return err
		} else {
			obj.State = schema.DeviceStateOffline
			content, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			_, _, err = tx.Set(key, string(content), nil)
			return err
		}
	})
}

func (d *Dao) DeleteDevice(deviceId string, userID string) error {
	key := schema.DeviceKey(userID, deviceId)
	return d.d.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		var obj *schema.Device
		if err = json.Unmarshal([]byte(val), &obj); err != nil {
			return err
		}
		if obj.State == schema.DeviceStateOffline {
			_, err = tx.Delete(key)
			return err
		} else {
			obj.Temp = true
			content, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			_, _, err = tx.Set(key, string(content), nil)
			return err
		}
	})
}

func (d *Dao) GetDevicesByUser(userID string, page, limit int) ([]*schema.Device, int64, error) {
	var total int64
	var list []*schema.Device
	err := d.d.View(func(tx *buntdb.Tx) error {
		pivot := fmt.Sprintf(`{"user":"%s"}`, userID)
		_ = tx.AscendEqual(schema.IdxDeviceUser, pivot, func(key, value string) bool {
			total++
			return true
		})

		skip := (page - 1) * limit
		var i int
		return tx.AscendEqual(schema.IdxDeviceUser, pivot, func(key, value string) bool {
			if i < skip {
				i++
				return true
			}
			if len(list) >= limit {
				return false
			}
			var obj *schema.Device
			if err := json.Unmarshal([]byte(value), &obj); err == nil {
				list = append(list, obj)
			}
			return true
		})
	})
	return list, total, err
}

func (d *Dao) UpdateDevice(userId, deviceId string, name *string, state schema.DeviceState) error {
	key := schema.DeviceKey(userId, deviceId)
	return d.d.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		var obj *schema.Device
		if err = json.Unmarshal([]byte(val), &obj); err != nil {
			return err
		}
		obj.Temp = false
		if name != nil {
			obj.Name = *name
		}
		if state != 0 {
			obj.State = state
		}
		content, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, string(content), nil)
		return err
	})
}
