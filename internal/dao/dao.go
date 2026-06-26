package dao

import (
	"os"
	"path/filepath"

	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/tidwall/buntdb"
)

type Dao struct {
	d *buntdb.DB
}

func (d *Dao) Close() error {
	return d.d.Close()
}

func (d *Dao) init() error {
	err := d.d.CreateIndex(schema.IdxUserName, schema.UserPattern, buntdb.IndexJSON("name"))
	if err != nil {
		return err
	}

	err = d.d.CreateIndex(schema.IdxInvitor, schema.InviteCodePattern, buntdb.IndexJSON("user"))
	if err != nil {
		return err
	}

	err = d.d.CreateIndex(schema.IdxDeviceUser, schema.DevicePattern, buntdb.IndexJSON("user"))
	if err != nil {
		return err
	}

	return nil
}

func New(path string) (bool, *Dao, error) {
	d := &Dao{}
	var init bool
	if path != ":memory:" {
		if _, err := os.Stat(path); err != nil {
			dir := filepath.Dir(path)
			_ = os.MkdirAll(dir, os.ModePerm)
			init = true
		}
	} else {
		init = true
	}

	db, err := buntdb.Open(path)
	if err != nil {
		return false, nil, err
	}
	d.d = db
	if err = d.init(); err != nil {
		_ = d.Close()
		return false, nil, err
	}
	return init, d, nil
}
