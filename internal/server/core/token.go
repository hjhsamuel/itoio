package core

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/hjhsamuel/itoio/internal/dao/schema"
)

func (c *Core) CheckRefreshToken(data string) (*schema.Token, error) {
	return c.d.CheckToken(data)
}

func (c *Core) RegenerateRefreshToken(userId, deviceId string) (string, error) {
	token, err := c.generateRefreshToken()
	if err != nil {
		return "", err
	}
	err = c.d.AddToken(&schema.Token{
		User:   userId,
		Device: deviceId,
		Data:   token,
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (c *Core) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
