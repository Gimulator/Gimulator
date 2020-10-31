package storage

import (
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
	GetUserWithToken(string) (*api.User, error)
	GetUserWithID(string) (*api.User, error)
	GetUsers(api.Character) ([]*api.User, error)
	UpdateUserStatus(string, api.Status) error
	UpdateUserReadiness(string, bool) error
}

type RoleStorage interface {
	GetRules(api.Character, string, api.Method) ([]*api.Key, error)
}
