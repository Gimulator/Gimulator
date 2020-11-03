package storage

import (
	"github.com/Gimulator/protobuf/go/api"
)

type MessageStorage interface {
	Put(message *api.Message) error
	Delete(key *api.Key) error
	DeleteAll(key *api.Key) error
	Get(key *api.Key) (*api.Message, error)
	GetAll(key *api.Key) ([]*api.Message, error)
}

type UserStorage interface {
	GetUsers(name *string, token *string, character *api.Character, role *string, readiness *bool, status *api.Status) ([]*api.User, error)
	GetUserWithToken(token string) (*api.User, error)
	GetUserWithName(name string) (*api.User, error)
	UpdateUserStatus(name string, status api.Status) error
	UpdateUserReadiness(name string, readiness bool) error
}

type RuleStorage interface {
	GetRules(character api.Character, role string, method api.Method) ([]*api.Key, error)
}
