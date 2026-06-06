package core

import (
	"fmt"
	"net"

	"github.com/hjhsamuel/itoio/config"
	"github.com/pion/turn/v5"
)

type TurnServer struct {
	l           net.PacketConn
	s           *turn.Server
	authHandler turn.AuthHandler // No authentication handler is set, the server is running in STUN only mode
	config      config.TurnConfig
}

func (s *TurnServer) SetAuthHandler(f turn.AuthHandler) {
	s.authHandler = f
}

func (s *TurnServer) Start() error {
	if s.config.PublicIP == "" {
		return fmt.Errorf("public ip is empty")
	}

	l, err := net.ListenPacket("udp4", fmt.Sprintf("0.0.0.0:%d", s.config.Port))
	if err != nil {
		return err
	}
	s.l = l

	turnConfig := turn.ServerConfig{
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: l,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(s.config.PublicIP),
					Address:      "0.0.0.0",
				},
			},
		},
		Realm:       s.config.Realm,
		AuthHandler: s.authHandler,
	}

	turnServer, err := turn.NewServer(turnConfig)
	if err != nil {
		_ = l.Close()
		return err
	}
	s.s = turnServer
	return nil
}

func (s *TurnServer) Close() error {
	if s.s != nil {
		return s.s.Close()
	}
	return nil
}

func NewTurnServer(c config.TurnConfig) *TurnServer {
	s := &TurnServer{
		authHandler: nil,
		config:      c,
	}
	return s
}
