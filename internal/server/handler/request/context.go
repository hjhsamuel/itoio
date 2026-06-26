package request

import (
	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
	User *User
}

type User struct {
	ID     string `json:"id"`
	Device string `json:"device"`
}

const (
	AuthCtxKey        = "AUTH"
	AuthCookieKey     = "ito"
	ConnTypeCookieKey = ""
)
