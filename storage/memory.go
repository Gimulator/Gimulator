package storage

import (
	"fmt"

	"github.com/Gimulator/protobuf/go/api"
)

type Identifier struct {
	Type      string
	Name      string
	Namespace string
}

//Casts a key type to an identifier type
func KeyToiden(key *api.Key) *Identifier {
	return &Identifier{
		Type:      key.Type,
		Name:      key.Name,
		Namespace: key.Namespace,
	}
}

type Memory struct {
	storage map[Identifier]*api.Message
}

func NewMemory() *Memory {
	return &Memory{
		storage: make(map[Identifier]*api.Message),
	}
}

func (m *Memory) Get(key *api.Key) (*api.Message, error) {
	iden := KeyToiden(key)
	return m.get(iden)
}

func (m *Memory) Put(msg *api.Message) error {
	return m.put(msg)
}

func (m *Memory) Delete(key *api.Key) error {
	iden := KeyToiden(key)
	return m.delete(iden)
}

func (m *Memory) GetAll(key *api.Key) ([]*api.Message, error) {
	iden := KeyToiden(key)
	return m.getall(iden), nil
}

func (iden *Identifier) MatchKeys(key *Identifier) bool {
	if iden.Type != "" && iden.Type != key.Type {
		return false
	} else if iden.Namespace != "" && iden.Namespace != key.Namespace {
		return false
	} else if iden.Name != "" && iden.Name != key.Name {
		return false
	}
	return true
}

func (m *Memory) get(iden *Identifier) (*api.Message, error) {
	if msg, exists := m.storage[*iden]; exists {
		return msg, nil
	}
	return nil, fmt.Errorf("object message with key=%v does not exist", iden)
}

//puts a message in storage
func (m *Memory) put(msg *api.Message) error {
	iden := KeyToiden(msg.Key)
	m.storage[*iden] = msg
	return nil
}

func (m *Memory) delete(iden *Identifier) error {
	if _, exists := m.storage[*iden]; exists {
		delete(m.storage, *iden)
		return nil
	}
	return fmt.Errorf("message object with key=%v does not exist", iden)
}

func (m *Memory) getall(iden *Identifier) []*api.Message {
	result := make([]*api.Message, 0)
	for i, o := range m.storage {
		if iden.MatchKeys(&i) {
			result = append(result, o)
		}
	}
	return result
}
