package core

import (
	"fmt"

	"github.com/hjhsamuel/itoio/config"
	"github.com/hjhsamuel/itoio/internal/entities"
)

type EventConnect struct{}

func (e *EventConnect) Execute(c *Core, info *ConnBase) error {
	if old := c.client[info.ID]; old != nil && old != info {
		c.cancelDisconnectTimer(info.ID)
	}
	c.client[info.ID] = info
	if roomObj, err := c.rooms.ReconnectUser(info.ID, info.Write); err == nil {
		c.cancelDisconnectTimer(info.ID)
		c.roomInfoChanged(roomObj)
	}
	Write(info.Write, c.iceConfig(info.ID))
	return nil
}

func (c *Core) iceConfig(userID string) *entities.IceConfig {
	host := c.turnConfig.PublicIP
	port := c.turnConfig.Port
	servers := []*entities.IceServer{}
	if host != "" && port > 0 {
		servers = append(servers, &entities.IceServer{
			URLs: []string{fmt.Sprintf("stun:%s:%d", host, port)},
		})
		if c.turnConfig.Mode == config.TurnModeTurn {
			username := turnUsername(userID)
			servers = append(servers, &entities.IceServer{
				URLs:       []string{fmt.Sprintf("turn:%s:%d", host, port)},
				Username:   username,
				Credential: turnCredential(username),
			})
		}
	}
	return &entities.IceConfig{
		IceServers: servers,
		Mode:       c.turnConfig.Mode,
	}
}
