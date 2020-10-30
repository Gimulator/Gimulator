package storage

import (
	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
)

type MessageStorage interface {
	Put(*api.Message) error
	Delete(*api.Key) error
	DeleteAll(*api.Key) error
	Get(*api.Key) (*api.Message, error)
	GetAll(*api.Key) ([]*api.Message, error)
}

type UserStorage interface {
	GetUserWithToken(string) (*User, error)
	GetUserWithID(string) (*User, error)
	UpdateUserStatus(string, types.Status) error
}

type RoleStorage interface {
	GetRules(string, types.Method) ([]*api.Key, error)
}

type User struct {
	ID     string
	Token  string
	Role   string
	Status types.Status
}
