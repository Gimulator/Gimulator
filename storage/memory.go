package storage

import (
	"fmt"

	"github.com/Gimulator/Gimulator/object"
)

type Memory struct {
	storage map[object.Key]*object.Object
}

func NewMemory() *Memory {
	return &Memory{
		storage: make(map[object.Key]*object.Object),
	}
}

func (m *Memory) Get(key *object.Key) (*object.Object, error) {
	return m.get(key)
}

func (m *Memory) Set(obj *object.Object) error {
	return m.set(obj)
}

func (m *Memory) Delete(key *object.Key) error {
	return m.delete(key)
}

func (m *Memory) Find(key *object.Key) ([]*object.Object, error) {
	return m.find(key), nil
}

func (m *Memory) get(key *object.Key) (*object.Object, error) {
	err := m.validateKey(key)
	if err != nil {
		return nil, err
	}

	if object, exists := m.storage[*key]; exists {
		return object, nil
	}
	return nil, fmt.Errorf("object with key=%v does not exist", key)
}

func (m *Memory) set(obj *object.Object) error {
	err := m.validateKey(obj.Key)
	if err != nil {
		return err
	}

	m.storage[*obj.Key] = obj
	return nil
}

func (m *Memory) delete(key *object.Key) error {
	err := m.validateKey(key)
	if err != nil {
		return err
	}

	if _, exists := m.storage[*key]; exists {
		delete(m.storage, *key)
		return nil
	}
	return fmt.Errorf("object with key=%v does not exist", key)
}

func (m *Memory) find(key *object.Key) []*object.Object {
	result := make([]*object.Object, 0)
	for k, o := range m.storage {
		if key.Match(&k) {
			result = append(result, o)
		}
	}
	return result
}

func (m *Memory) validateKey(key *object.Key) error {
	if key.Name == "" {
		return fmt.Errorf("invalid key with empty Name")
	}
	if key.Namespace == "" {
		return fmt.Errorf("invalid key with empty Namespace")
	}
	if key.Type == "" {
		return fmt.Errorf("invalid key with empty Type")
	}
	return nil
}
