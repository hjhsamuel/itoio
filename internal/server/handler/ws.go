package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hjhsamuel/itoio/internal/server/handler/request"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (a *api) WebSocket(ctx *gin.Context) {
	val, exists := ctx.Get(request.AuthCtxKey)
	if !exists {
		return
	}
	user, ok := val.(*request.User)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	a.srv.WebSocketHandler(user, conn)
}
