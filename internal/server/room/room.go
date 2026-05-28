package room

import (
	"errors"

	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/pkg/password"
)

type Room struct {
	ID                string
	Secret            string
	CloseOnOwnerLeave bool
	Users             map[string]*RoomUser
}

func (r *Room) GetUsers() []*RoomUser {
	users := make([]*RoomUser, 0, len(r.Users))
	for _, user := range r.Users {
		users = append(users, user)
	}
	return users
}

func (r *Room) AddWatcher(secret string, user *RoomUser) error {
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

func NewRoom(id, secret string, owner *RoomUser) (*Room, error) {
	r := &Room{
		ID:                id,
		CloseOnOwnerLeave: true,
		Users:             make(map[string]*RoomUser),
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
