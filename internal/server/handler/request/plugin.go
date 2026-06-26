package request

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/common"
	"github.com/hjhsamuel/itoio/pkg/token"
)

func WithAuth(ctx *gin.Context) {
	cookie, err := ctx.Cookie(AuthCookieKey)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, &Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	claim, err := token.Decode(cookie, common.JwtSalt)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, &Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	ctx.Set(AuthCtxKey, &User{
		ID:     claim.ID,
		Device: claim.Device,
	})
	ctx.Next()
}
