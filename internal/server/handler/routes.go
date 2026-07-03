package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/server/core"
	"github.com/hjhsamuel/itoio/internal/server/handler/request"
	"github.com/hjhsamuel/itoio/ui"
)

type api struct {
	srv *core.Core
}

func Routes(c *core.Core, e *gin.Engine) error {
	ui.Route(e)

	a := &api{srv: c}

	e.POST("/login", a.Login)
	e.POST("/register", a.Register)
	e.PUT("/refresh", a.Refresh)

	group := e.Group("/ito")
	group.Use(request.WithAuth)
	group.GET("/stream", a.WebSocket)
	group.GET("/admin/code", a.GetInviteCodeList)
	group.POST("/admin/code", a.CreateInviteCode)
	group.PUT("/admin/passwd", a.UpdatePassword)
	group.POST("/device", a.CreateDevice)
	group.PUT("/device", a.UpdateDevice)
	group.DELETE("/device", a.DelDevice)
	group.GET("/device", a.MineDevices)
	group.PUT("/room", a.UpdateRoomConfig)

	return nil
}
