package storage

import (
	"fmt"

	"gitlab.com/Syfract/Xerac/gimulator/object"
)

type Memory struct {
	storage map[object.Key]object.Object
}

func NewMemory() *Memory {
	return &Memory{
		storage: make(map[object.Key]object.Object),
	}
}

func (m *Memory) Get(key object.Key) (object.Object, error) {
	return m.get(key)
}

func (m *Memory) Set(obj object.Object) error {
	m.set(obj)
	return nil
}

func (m *Memory) Delete(key object.Key) error {
	return m.del(key)
}

func (m *Memory) Find(key object.Key) ([]object.Object, error) {
	return m.find(key), nil
}

func (m *Memory) get(key object.Key) (object.Object, error) {
	if object, exists := m.storage[key]; exists {
		return object, nil
	}
	return object.Object{}, fmt.Errorf("object with %v key does not exist", key)
}

func (m *Memory) set(obj object.Object) {
	m.storage[obj.Key] = obj
}

func (m *Memory) del(key object.Key) error {
	if _, exists := m.storage[key]; exists {
		delete(m.storage, key)
		return nil
	}
	return fmt.Errorf("object with %v key does not exist", key)
}

func (m *Memory) find(key object.Key) []object.Object {
	result := make([]object.Object, 0)
	for k, o := range m.storage {
		if k.Match(key) {
			result = append(result, o)
		}
	}
	return result
}
