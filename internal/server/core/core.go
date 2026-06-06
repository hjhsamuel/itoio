package core

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hjhsamuel/itoio/config"
	"github.com/hjhsamuel/itoio/internal/common"
	"github.com/hjhsamuel/itoio/internal/dao"
	"github.com/hjhsamuel/itoio/internal/server/room"
	"github.com/pion/turn/v5"
)

type Core struct {
	id int64

	d *dao.Dao

	rooms            *room.RoomManager
	client           map[string]*ConnBase
	disconnectTimers map[string]*time.Timer
	turn             *TurnServer
	turnConfig       config.TurnConfig

	read chan *ConnMessage

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *Core) Start(conf *config.Config) error {
	c.turnConfig = conf.Turn
	c.turn = NewTurnServer(conf.Turn)
	if conf.Turn.Mode == config.TurnModeTurn {
		c.turn.SetAuthHandler(c.turnAuthHandler)
	}
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

func (c *Core) turnAuthHandler(ra *turn.RequestAttributes) (string, []byte, bool) {
	expire, userID, ok := strings.Cut(ra.Username, ":")
	if !ok {
		return "", nil, false
	}
	expireAt, err := strconv.ParseInt(expire, 10, 64)
	if err != nil || time.Now().Unix() > expireAt {
		return "", nil, false
	}
	if c.d == nil {
		return "", nil, false
	}
	if _, err = c.d.GetUserByID(userID); err != nil {
		return "", nil, false
	}
	credential := turnCredential(ra.Username)
	key := turn.GenerateAuthKey(ra.Username, c.turnConfig.Realm, credential)
	return ra.Username, key, true
}

func turnCredential(username string) string {
	mac := hmac.New(sha1.New, []byte(common.JwtSalt))
	_, _ = mac.Write([]byte(username))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func turnUsername(userID string) string {
	return fmt.Sprintf("%d:%s", time.Now().Add(common.JwtExp).Unix(), userID)
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
