package storage

import (
	"fmt"

	"github.com/Gimulator/protobuf/go/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type identifier struct {
	theType   string
	name      string
	namespace string
}

//Casts a key type to an identifier type
func keyToiden(key *api.Key) *identifier {
	return &identifier{
		theType:   key.Type,
		name:      key.Name,
		namespace: key.Namespace,
	}
}

type Memory struct {
	storage map[identifier]*api.Message
}

func NewMemory() *Memory {
	return &Memory{
		storage: make(map[identifier]*api.Message),
	}
}

func (m *Memory) Get(key *api.Key) (*api.Message, error) {
	iden := keyToiden(key)
	getMsgResult, err := m.get(iden)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not get any message for key=%v", iden))
	}
	return getMsgResult, nil
}

func (m *Memory) get(iden *identifier) (*api.Message, error) {
	if msg, exists := m.storage[*iden]; exists {
		return msg, nil
	}
	return nil, fmt.Errorf("object message with key=%v does not exist", iden)
}

//puts a message in storage
func (m *Memory) Put(msg *api.Message) error {
	putMsgResult := m.put(msg)
	if putMsgResult != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not put message=%v in storage", msg))
	}
	return putMsgResult
}

func (m *Memory) put(msg *api.Message) error {
	iden := keyToiden(msg.Key)
	m.storage[*iden] = msg
	return nil
}

func (m *Memory) Delete(key *api.Key) error {
	iden := keyToiden(key)
	delMsgResult := m.delete(iden)
	if delMsgResult != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not delete message with key=%v ", iden))
	}
	return delMsgResult
}

func (m *Memory) delete(iden *identifier) error {
	if _, exists := m.storage[*iden]; exists {
		delete(m.storage, *iden)
		return nil
	}
	return fmt.Errorf("message object with key=%v does not exist", iden)
}

func (m *Memory) GetAll(key *api.Key) ([]*api.Message, error) {
	iden := keyToiden(key)
	getallMsgsResult, err := m.getall(iden)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not getall messages with key=%v ", iden))
	}
	return getallMsgsResult, nil
}

func (m *Memory) getall(iden *identifier) ([]*api.Message, error) {
	result := make([]*api.Message, 0)
	for i, o := range m.storage {
		if iden.matchKeys(&i) {
			result = append(result, o)
		}
	}

	if result != nil {
		return result, nil
	}

	return nil, fmt.Errorf("no messages matched with key=%v ", iden)
}

func (iden *identifier) matchKeys(key *identifier) bool {
	if iden.theType != "" && iden.theType != key.theType {
		return false
	} else if iden.namespace != "" && iden.namespace != key.namespace {
		return false
	} else if iden.name != "" && iden.name != key.name {
		return false
	}
	return true
}
