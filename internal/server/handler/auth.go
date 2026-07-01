package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/server/core"
	"github.com/hjhsamuel/itoio/internal/server/handler/request"
	"github.com/hjhsamuel/itoio/internal/server/handler/schema"
)

func (a *api) Login(ctx *gin.Context) {
	var req *schema.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "invalid username or password",
		})
		return
	}
	user, expire, token, err := a.srv.UserLogin(req.Name, req.Password, req.Device)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusInternalServerError,
			Message: "invalid username or password",
		})
		return
	}

	if req.Device != "" && req.DevName != "" {
		err = a.srv.AddDevice(user.ID, req.Device, req.DevName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, &request.Response{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
	}

	ctx.SetCookie(request.AuthCookieKey, token, expire, "/", "", false, true)
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
		Data: &schema.LoginRsp{
			ID:       core.GenerateConnUUID(user.ID, req.Device),
			Nickname: user.Nickname,
			Expire:   expire,
		},
	})
}

func (a *api) Register(ctx *gin.Context) {
	var req *schema.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}
	if req.Name == "" || req.Password == "" || req.Nickname == "" {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}
	_, err := a.srv.UserRegister(req.Code, req.Name, req.Password, req.Nickname)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
	})
}

func (a *api) Refresh(ctx *gin.Context) {
	val, exists := ctx.Get(request.AuthCtxKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	expire, token, err := a.srv.RefreshToken(user.ID, user.Device)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	ctx.SetCookie(request.AuthCookieKey, token, expire, "/", "", false, true)
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
		Data: &schema.RefreshRsp{
			Expire: expire,
		},
	})
}
