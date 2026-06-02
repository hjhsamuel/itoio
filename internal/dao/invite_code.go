package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/hjhsamuel/itoio/pkg/random"
	"github.com/tidwall/buntdb"
)

func (d *Dao) GetInviteCodeCnt(userId string) (int64, error) {
	var cnt int64
	err := d.d.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(schema.IdxInvitor, fmt.Sprintf(`{"user":"%s"}`, userId), func(key, value string) bool {
			cnt += 1
			return true
		})
	})
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func (d *Dao) CreateInviteCode(nodeId int64, userId string, expire time.Duration) (string, int64, error) {
	var (
		now      = time.Now()
		expireAt = now.Add(expire)
		salt     = rand.Intn(10000)
	)
	data := fmt.Sprintf("%d:%s:%d:%d:%d", nodeId, userId, now.Unix(), expireAt.Unix(), salt)
	code, err := random.Generate62Str(data)
	if err != nil {
		return "", 0, err
	}
	obj := &schema.InviteCode{
		Code:   code,
		User:   userId,
		Expire: expireAt.Unix(),
	}
	err = d.d.Update(func(tx *buntdb.Tx) error {
		key := schema.InviteCodeKey(code)
		val, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		_, replaced, err := tx.Set(key, string(val), &buntdb.SetOptions{
			Expires: true,
			TTL:     expire,
		})
		if replaced || err != nil {
			return errors.New("retry later")
		}
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return code, obj.Expire, nil
}

func (d *Dao) CheckInviteCode(code string) bool {
	key := schema.InviteCodeKey(code)
	err := d.d.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(key)
		return err
	})
	if err != nil {
		return false
	}
	return true
}

func (d *Dao) GetInviteCodeList(userId string, offset, limit int) ([]*schema.InviteCode, error) {
	vals := make([]string, 0, limit)
	err := d.d.View(func(tx *buntdb.Tx) error {
		var (
			skip = 0
			cnt  = 0
		)
		return tx.AscendEqual(schema.IdxInvitor, fmt.Sprintf(`{"user":"%s"}`, userId), func(key, value string) bool {
			if cnt >= limit {
				return false
			}
			if skip < offset {
				skip += 1
				return true
			}
			vals = append(vals, value)
			cnt += 1
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	res := make([]*schema.InviteCode, 0, limit)
	for _, val := range vals {
		obj := &schema.InviteCode{}
		if err = json.Unmarshal([]byte(val), obj); err != nil {
			return nil, err
		}
		res = append(res, obj)
	}
	return res, nil
}

func (d *Dao) DelInviteCode(userId, code string) error {
	key := schema.InviteCodeKey(code)
	return d.d.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Delete(key)
		if err != nil {
			return err
		}
		var obj *schema.InviteCode
		if err = json.Unmarshal([]byte(val), &obj); err != nil {
			return err
		}
		if obj.User != userId {
			return errors.New("not your invite code")
		}
		return nil
	})
}
