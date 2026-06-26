package room

import (
	"errors"
	"time"

	"github.com/hjhsamuel/itoio/internal/common"
	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/pkg/snowflake"
)

type RoomManager struct {
	Rooms     map[string]*Room
	connected map[string]string
	gen       *snowflake.Node
}

func (m *RoomManager) GetCurrentRoom(id string) (*Room, error) {
	roomId, ok := m.connected[id]
	if !ok {
		return nil, errors.New("user not in any room")
	}
	room, ok := m.Rooms[roomId]
	if !ok {
		return nil, errors.New("room not found")
	}
	return room, nil
}

func (m *RoomManager) CreateRoom(secret string, mode RoomMode, user *RoomUser) (string, error) {
	if _, ok := m.connected[user.ID]; ok {
		return "", errors.New("user already in a room")
	}

	var prefix string
	if mode == RoomModeControl {
		prefix = "rd-"
	}

	var roomId string
	for i := 0; i < 3; i++ {
		id := m.gen.Generate()
		rid := prefix + id.Base58Unpredictable([]byte(common.RoomIdSalt))
		if _, ok := m.Rooms[rid]; ok {
			time.Sleep(10 * time.Microsecond)
			continue
		}
		roomId = rid
		break
	}
	if roomId == "" {
		return "", errors.New("create room failed, please try again later")
	}

	room, err := NewRoom(roomId, secret, mode, user)
	if err != nil {
		return "", err
	}
	m.Rooms[roomId] = room
	m.connected[user.ID] = roomId
	return roomId, nil
}

func (m *RoomManager) CloseRoom(roomId string) {
	room, ok := m.Rooms[roomId]
	if !ok {
		return
	}
	users := room.GetUsers()
	for _, user := range users {
		delete(m.connected, user.ID)
	}
	delete(m.Rooms, roomId)
}

func (m *RoomManager) JoinRoom(roomId, secret string, user *RoomUser) (*Room, error) {
	if _, ok := m.connected[user.ID]; ok {
		return nil, errors.New("user already in a room")
	}
	room, ok := m.Rooms[roomId]
	if !ok {
		return nil, errors.New("room not found")
	}
	err := room.AddWatcher(secret, user)
	if err != nil {
		return nil, err
	}
	m.connected[user.ID] = roomId
	return room, nil
}

func (m *RoomManager) ReconnectUser(userId string, write chan entities.Message) (*Room, error) {
	room, err := m.GetCurrentRoom(userId)
	if err != nil {
		return nil, err
	}
	if !room.SetUserWrite(userId, write) {
		return nil, errors.New("user not in room")
	}
	return room, nil
}

func (m *RoomManager) MarkDisconnected(userId string) (*Room, error) {
	return m.ReconnectUser(userId, nil)
}

func (m *RoomManager) LeaveRoom(userId string) bool {
	roomId, ok := m.connected[userId]
	if !ok {
		return false
	}
	delete(m.connected, userId)
	room, ok := m.Rooms[roomId]
	if !ok {
		return false
	}
	return room.RemoveWatcher(userId)
}

func (m *RoomManager) InSameRoom(ids ...string) bool {
	var roomId string
	for _, id := range ids {
		if v, ok := m.connected[id]; !ok {
			return false
		} else {
			if roomId == "" {
				roomId = v
			} else {
				if roomId != v {
					return false
				}
			}
		}
	}
	return true
}

func (m *RoomManager) UpdateSecret(roomId, userId, secret string) error {
	room, ok := m.Rooms[roomId]
	if !ok {
		return errors.New("room not found")
	}
	return room.UpdateSecret(userId, secret)
}

func NewRoomManager(node int64) (*RoomManager, error) {
	gen, err := snowflake.NewNode(node)
	if err != nil {
		return nil, err
	}

	m := &RoomManager{
		Rooms:     make(map[string]*Room),
		connected: make(map[string]string),
		gen:       gen,
	}
	return m, nil
}
