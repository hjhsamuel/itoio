package request

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/common"
	"github.com/hjhsamuel/itoio/pkg/token"
)

func WithAuth(ctx *gin.Context) {
	user, err := ParseCookie(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, &Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	ctx.Set(AuthCtxKey, user)
	ctx.Next()
}

func ParseCookie(ctx *gin.Context) (*User, error) {
	cookie, err := ctx.Cookie(AuthCookieKey)
	if err != nil {
		return nil, err
	}

	claim, err := token.Decode(cookie, common.JwtSalt)
	if err != nil {
		return nil, err
	}

	return &User{ID: claim.ID, Device: claim.Device}, nil
}
