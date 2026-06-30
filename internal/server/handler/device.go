package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/server/handler/request"
	"github.com/hjhsamuel/itoio/internal/server/handler/schema"
)

func (a *api) UpdateDevice(ctx *gin.Context) {
	val, _ := ctx.Get(request.AuthCtxKey)
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	var req schema.UpdateDeviceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}

	err := a.srv.UpdateDevice(user.ID, user.Device, req.Name, 0)
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

func (a *api) DelDevice(ctx *gin.Context) {
	val, _ := ctx.Get(request.AuthCtxKey)
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	var req schema.DelDeviceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}

	err := a.srv.DeleteDevice(req.ID, user.ID)
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

func (a *api) MineDevices(ctx *gin.Context) {
	val, _ := ctx.Get(request.AuthCtxKey)
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	var req schema.MineDevicesReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	devices, total, err := a.srv.GetDevices(user.ID, req.Page, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &request.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	rsp := schema.MineDeviceRsp{
		Total: total,
		Data:  make([]*schema.DeviceInfo, 0, len(devices)),
	}
	for _, d := range devices {
		rsp.Data = append(rsp.Data, &schema.DeviceInfo{
			ID:    d.ID,
			Name:  d.Name,
			State: d.State,
			Temp:  d.Temp,
			Room:  d.Room,
		})
	}

	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
		Data: rsp,
	})
}
