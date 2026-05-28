package dao

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/tidwall/buntdb"
)

func (d *Dao) CreateUser(obj *schema.User) error {
	id, err := d.autoIncr(schema.AutoIncrUser)
	if err != nil {
		return err
	}
	obj.ID = id
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return d.d.Update(func(tx *buntdb.Tx) error {
		var matched bool
		err = tx.AscendEqual(schema.IdxUserName, fmt.Sprintf(`{"name":"%s"}`, obj.Name), func(key, value string) bool {
			matched = true
			return true
		})
		if err != nil {
			return err
		}
		if matched {
			return errors.New("user already exists")
		}

		_, _, err = tx.Set(schema.UserKey(id), string(content), nil)
		return err
	})
}

func (d *Dao) GetUserByName(name string) (*schema.User, error) {
	var val string
	err := d.d.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(schema.IdxUserName, fmt.Sprintf(`{"name":"%s"}`, name), func(key, value string) bool {
			val = value
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	var obj *schema.User
	if err = json.Unmarshal([]byte(val), &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (d *Dao) GetUserByID(id string) (*schema.User, error) {
	key := schema.UserKey(id)
	var val string
	err := d.d.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get(key)
		if err != nil {
			return err
		}
		val = value
		return nil
	})
	if err != nil {
		return nil, err
	}
	var obj *schema.User
	if err = json.Unmarshal([]byte(val), &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (d *Dao) UpdateNickname(id string, nickname string) error {
	key := schema.UserKey(id)
	return d.d.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		var obj *schema.User
		err = json.Unmarshal([]byte(val), &obj)
		if err != nil {
			return err
		}
		obj.Nickname = nickname
		content, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, string(content), nil)
		return err
	})
}

func (d *Dao) UpdatePassword(id string, passwd string) error {
	key := schema.UserKey(id)
	return d.d.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		var obj *schema.User
		err = json.Unmarshal([]byte(val), &obj)
		if err != nil {
			return err
		}
		obj.Password = passwd
		content, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, string(content), nil)
		return err
	})
}
