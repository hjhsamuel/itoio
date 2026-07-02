package dao

import (
	"encoding/json"
	"fmt"

	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/tidwall/buntdb"
)

func (d *Dao) AddToken(obj *schema.Token) error {
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	key := schema.TokenKey(obj.User, obj.Device)
	return d.d.Update(func(tx *buntdb.Tx) error {
		_, _, err = tx.Set(key, string(content), nil)
		return err
	})
}

func (d *Dao) GetToken(user, device string) (*schema.Token, error) {
	key := schema.TokenKey(user, device)
	var obj *schema.Token
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

func (d *Dao) CheckToken(data string) (*schema.Token, error) {
	pivot := fmt.Sprintf(`{"data":"%s"}`, data)
	var obj *schema.Token
	err := d.d.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(schema.IdxTokenData, pivot, func(key, value string) bool {
			err := json.Unmarshal([]byte(value), &obj)
			if err != nil {
				return false
			}
			return true
		})
	})
	return obj, err
}
