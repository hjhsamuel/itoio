package dao

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/tidwall/buntdb"
)

func (d *Dao) CreateDevice(obj *schema.Device) error {
	var (
		content []byte
	)
	key := schema.DeviceKey(obj.User, obj.Device)
	return d.d.Update(func(tx *buntdb.Tx) error {
		// check if device exists
		val, err := tx.Get(key)
		if err != nil {
			if !errors.Is(err, buntdb.ErrNotFound) {
				return err
			}
			// first time or temporary device, add
			content, err = json.Marshal(obj)
			if err != nil {
				return err
			}
		} else {
			// set state to online
			var devObj *schema.Device
			if err = json.Unmarshal([]byte(val), &devObj); err != nil {
				return err
			}
			devObj.State = schema.DeviceStateOnline
			content, err = json.Marshal(devObj)
			if err != nil {
				return err
			}
		}
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

func (d *Dao) GetDevicesByUser(userID string) ([]*schema.Device, int64, error) {
	var (
		total  int64
		list   []*schema.Device
		delKey []string
	)
	err := d.d.View(func(tx *buntdb.Tx) error {
		pivot := fmt.Sprintf(`{"user":"%s"}`, userID)
		return tx.AscendEqual(schema.IdxDeviceUser, pivot, func(key, value string) bool {
			var obj *schema.Device
			if err := json.Unmarshal([]byte(value), &obj); err != nil {
				return true
			}
			switch obj.State {
			case schema.DeviceStateOnline:
				list = append(list, obj)
				total += 1
			case schema.DeviceStateOffline:
				if obj.Temp {
					delKey = append(delKey, key)
				} else {
					list = append(list, obj)
					total += 1
				}
			}
			return true
		})
	})
	if err != nil {
		return nil, 0, err
	}
	// remove device
	if len(delKey) != 0 {
		_ = d.d.Update(func(tx *buntdb.Tx) error {
			for _, key := range delKey {
				_, _ = tx.Delete(key)
			}
			return nil
		})
	}
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

func (d *Dao) GetDevice(userId, deviceId string) (*schema.Device, error) {
	var obj *schema.Device
	key := schema.DeviceKey(userId, deviceId)
	err := d.d.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &obj)
	})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
