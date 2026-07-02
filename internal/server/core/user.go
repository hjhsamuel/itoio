package core

import (
	"errors"
	"time"

	"github.com/hjhsamuel/itoio/internal/common"
	"github.com/hjhsamuel/itoio/internal/dao/schema"
	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/pkg/password"
	"github.com/hjhsamuel/itoio/pkg/token"
)

func (c *Core) InitFirstUser() (string, string, error) {
	passwd, err := password.Random(12)
	if err != nil {
		return "", "", err
	}
	out, err := password.HashPassword(passwd, password.DefaultParams)
	if err != nil {
		return "", "", err
	}
	err = c.d.CreateUser(&schema.User{
		Name:     "admin",
		Nickname: "admin",
		Password: out,
	})
	if err != nil {
		return "", "", err
	}
	return "admin", passwd, nil
}

func (c *Core) UserLogin(name, passwd, device string) (*schema.User, *entities.Login, error) {
	user, err := c.d.GetUserByName(name)
	if err != nil {
		return nil, nil, err
	}
	matched, err := password.VerifyPassword(passwd, user.Password)
	if err != nil || !matched {
		return nil, nil, errors.New("invalid username or password")
	}
	out, err := token.Encode(user.ID, device, common.JwtSalt, common.JwtExp)
	if err != nil {
		return nil, nil, err
	}

	t := &entities.Login{
		Token:  out,
		Expire: int64(common.JwtExp.Seconds()),
	}

	if device != "" {
		rt, err := c.generateRefreshToken()
		if err != nil {
			return nil, nil, err
		}
		t.RefreshToken = rt

		err = c.d.AddToken(&schema.Token{
			User:   user.ID,
			Device: device,
			Data:   rt,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return user, t, nil
}

func (c *Core) UserRegister(code string, name, passwd, nickname string) (string, error) {
	if !c.d.CheckInviteCode(code) {
		return "", errors.New("invalid code")
	}

	if err := password.Validator(passwd); err != nil {
		return "", err
	}
	out, err := password.HashPassword(passwd, password.DefaultParams)
	if err != nil {
		return "", err
	}
	user := &schema.User{
		Name:     name,
		Nickname: nickname,
		Password: out,
	}
	if err = c.d.CreateUser(user); err != nil {
		return "", err
	}
	return user.ID, nil
}

func (c *Core) RefreshToken(id, device string) (int, string, error) {
	_, err := c.d.GetUserByID(id)
	if err != nil {
		return 0, "", err
	}
	out, err := token.Encode(id, device, common.JwtSalt, common.JwtExp)
	if err != nil {
		return 0, "", err
	}
	return int(common.JwtExp.Seconds()), out, nil
}

func (c *Core) GetInviteCodeList(userId string, offset, limit int) ([]*schema.InviteCode, error) {
	return c.d.GetInviteCodeList(userId, offset, limit)
}

func (c *Core) CreateInviteCode(userId string) (string, int64, error) {
	cnt, err := c.d.GetInviteCodeCnt(userId)
	if err != nil {
		return "", 0, errors.New("check invite code failed")
	}
	if cnt >= 5 {
		return "", 0, errors.New("reached max invite code limit")
	}
	return c.d.CreateInviteCode(c.id, userId, time.Hour*24*7)
}

func (c *Core) UpdatePasswd(userId, oldPasswd, newPasswd string) error {
	if oldPasswd == newPasswd {
		return errors.New("password should not be the same")
	}
	if err := password.Validator(newPasswd); err != nil {
		return err
	}
	user, err := c.d.GetUserByID(userId)
	if err != nil {
		return errors.New("invalid request")
	}
	matched, err := password.VerifyPassword(oldPasswd, user.Password)
	if err != nil || !matched {
		return errors.New("invalid password")
	}
	out, err := password.HashPassword(newPasswd, password.DefaultParams)
	if err != nil {
		return err
	}
	return c.d.UpdatePassword(userId, out)
}
