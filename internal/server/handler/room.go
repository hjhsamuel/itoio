package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/server/handler/request"
	"github.com/hjhsamuel/itoio/internal/server/handler/schema"
)

func (a *api) UpdateRoomConfig(ctx *gin.Context) {
	val, _ := ctx.Get(request.AuthCtxKey)
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	var req schema.UpdateRoomConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}
	err := a.srv.UpdateRoomConfig(req.Room, user.ID, user.Device, req.Secret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &request.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
	})
}
