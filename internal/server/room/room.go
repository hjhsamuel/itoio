package room

import (
	"errors"
	"strings"

	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/pkg/password"
)

type RoomMode int

const (
	RoomModeLive    RoomMode = iota + 1 // live
	RoomModeControl                     // control
)

type Room struct {
	ID     string
	Secret string
	Mode   RoomMode
	Owner  string
	Users  map[string]*RoomUser
}

func (r *Room) GetUsers() []*RoomUser {
	users := make([]*RoomUser, 0, len(r.Users))
	for _, user := range r.Users {
		users = append(users, user)
	}
	return users
}

func (r *Room) AddWatcher(secret string, user *RoomUser) error {
	if r.Mode == RoomModeControl {
		// owner can join without secret
		parts := strings.FieldsFunc(user.ID, func(r rune) bool {
			return r == ':'
		})
		if strings.HasPrefix(r.Owner, parts[0]+":") {
			r.Users[user.ID] = user
			return nil
		}
	}

	if r.Secret != "" {
		matched, err := password.VerifyPassword(secret, r.Secret)
		if err != nil {
			return err
		}
		if !matched {
			return errors.New("password is incorrect")
		}
	}
	r.Users[user.ID] = user
	return nil
}

func (r *Room) UpdateSecret(id string, secret string) error {
	if val, ok := r.Users[id]; ok && val.Owner {
		passwd, err := password.HashPassword(secret, password.DefaultParams)
		if err != nil {
			return err
		}
		r.Secret = passwd
		return nil
	}
	return errors.New("permission denied")
}

func (r *Room) SetUserWrite(id string, write chan entities.Message) bool {
	user, ok := r.Users[id]
	if !ok {
		return false
	}
	user.Write = write
	return true
}

func (r *Room) RemoveWatcher(id string) bool {
	user, ok := r.Users[id]
	if !ok {
		return false
	}
	delete(r.Users, id)
	if user.Owner {
		return true
	}
	return false
}

func NewRoom(id, secret string, mode RoomMode, owner *RoomUser) (*Room, error) {
	r := &Room{
		ID:    id,
		Mode:  mode,
		Users: make(map[string]*RoomUser),
		Owner: owner.ID,
	}
	if secret != "" {
		passwd, err := password.HashPassword(secret, password.DefaultParams)
		if err != nil {
			return nil, err
		}
		r.Secret = passwd
	}
	r.Users[owner.ID] = owner
	return r, nil
}

type RoomUser struct {
	ID       string
	Nickname string
	Owner    bool
	Write    chan entities.Message
}
