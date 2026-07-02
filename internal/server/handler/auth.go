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

	var deviceId string
	if req.Device != nil {
		deviceId = req.Device.ID
	}

	user, token, err := a.srv.UserLogin(req.Name, req.Password, deviceId)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusInternalServerError,
			Message: "invalid username or password",
		})
		return
	}

	if deviceId != "" {
		err = a.srv.AddDevice(user.ID, req.Device.ID, req.Device.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, &request.Response{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
	}

	ctx.SetCookie(request.AuthCookieKey, token.Token, int(token.Expire), "/", "", false, true)
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
		Data: &schema.LoginRsp{
			ID:       core.GenerateConnUUID(user.ID, deviceId),
			Nickname: user.Nickname,
			Expire:   int(token.Expire),
			Token:    token.RefreshToken,
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
	var req *schema.RefreshReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}

	var (
		userId   string
		deviceId string
	)
	if req.Token == "" {
		// web
		user, err := request.ParseCookie(ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, &request.Response{
				Code:    http.StatusUnauthorized,
				Message: "unauthorized",
			})
			return
		}
		userId, deviceId = user.ID, user.Device
	} else {
		// agent
		obj, err := a.srv.CheckRefreshToken(req.Token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, &request.Response{
				Code:    http.StatusUnauthorized,
				Message: "unauthorized",
			})
			return
		}
		userId, deviceId = obj.User, obj.Device
	}

	expire, token, err := a.srv.RefreshToken(userId, deviceId)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	rsp := &schema.RefreshRsp{
		Expire: expire,
	}
	if req.Token != "" {
		refreshToken, err := a.srv.RegenerateRefreshToken(userId, deviceId)
		if err == nil {
			rsp.Token = refreshToken
		}
	}

	ctx.SetCookie(request.AuthCookieKey, token, expire, "/", "", false, true)
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
		Data: rsp,
	})
}
