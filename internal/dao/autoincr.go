package dao

import (
	"errors"
	"strconv"

	"github.com/tidwall/buntdb"
)

func (d *Dao) autoIncr(key string) (string, error) {
	var id string
	err := d.d.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			if errors.Is(err, buntdb.ErrNotFound) {
				id = "1"
				_, _, err = tx.Set(key, "1", nil)
				return err
			}
			return err
		}
		x, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		x += 1
		id = strconv.FormatUint(x, 10)
		_, _, err = tx.Set(key, id, nil)
		return err
	})
	if err != nil {
		return "", err
	}
	return id, nil
}
