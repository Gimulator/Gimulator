package storage

import "github.com/Gimulator/Gimulator/object"

type Storage interface {
	Set(*object.Object) error
	Delete(*object.Key) error
	Get(*object.Key) (*object.Object, error)
	Find(*object.Key) ([]*object.Object, error)
}
