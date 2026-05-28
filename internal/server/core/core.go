package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hjhsamuel/itoio/config"
	"github.com/hjhsamuel/itoio/internal/dao"
	"github.com/hjhsamuel/itoio/internal/server/core/turn"
	"github.com/hjhsamuel/itoio/internal/server/room"
)

type Core struct {
	id int64

	d *dao.Dao

	rooms            *room.RoomManager
	client           map[string]*ConnBase
	disconnectTimers map[string]*time.Timer
	turn             *turn.TurnServer
	turnConfig       config.TurnConfig

	read chan *ConnMessage

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *Core) Start(conf *config.Config) error {
	c.turnConfig = conf.Turn
	c.turn = turn.NewTurnServer(conf.Turn)
	if err := c.turn.Start(); err != nil {
		return err
	}

	init, d, err := dao.New(conf.Server.StoragePath)
	if err != nil {
		return err
	}
	c.d = d

	if init {
		name, passwd, err := c.InitFirstUser()
		if err != nil {
			return err
		}
		fmt.Printf("Initialized first user: %s/%s\n", name, passwd)
	}

	c.wg.Add(1)
	go c.do()
	return nil
}

func (c *Core) Close() error {
	if c.turn != nil {
		_ = c.turn.Close()
	}
	if c.d != nil {
		_ = c.d.Close()
	}
	c.cancel()
	c.wg.Wait()
	return nil
}

func NewCore(id int64) (*Core, error) {
	rooms, err := room.NewRoomManager(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Core{
		id:               id,
		rooms:            rooms,
		client:           make(map[string]*ConnBase),
		disconnectTimers: make(map[string]*time.Timer),
		read:             make(chan *ConnMessage),
		ctx:              ctx,
		cancel:           cancel,
	}
	return s, nil
}
