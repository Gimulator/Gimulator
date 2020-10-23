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

type CredentialStorage interface {
	GetCredWithToken(string) (string, string, error)
}

type RolesStorage interface {
	GetRules(string, types.Method) ([]*api.Key, error)
}

type AuthStorage interface {
	CredentialStorage
	RolesStorage
}
