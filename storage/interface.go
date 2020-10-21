package storage

import "github.com/Gimulator/protobuf/go/api"

type Storage interface {
	Put(*api.Message) error
	Delete(*api.Key) error
	DeleteAll(*api.Key) error
	Get(*api.Key) (*api.Message, error)
	GetAll(*api.Key) ([]*api.Message, error)
}
