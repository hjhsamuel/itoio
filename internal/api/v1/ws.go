package v1

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hjhsamuel/itoio/conf"
)

func ws(a *api, g *gin.RouterGroup) {
	g.GET("/ws", a.handleWebSocket)
}

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("origin")
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		if u.Host == r.Host {
			return true
		}
		return conf.Conf.CheckOrigin(origin)
	},
}

func (a *api) handleWebSocket(c *gin.Context) {
	conn, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
}
