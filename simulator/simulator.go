package simulator

import (
	"sync"

	"github.com/Gimulator/Gimulator/storage"
	"github.com/Gimulator/protobuf/go/api"
)

type Simulator struct {
	sync.Mutex
	spreader *spreader
	storage  storage.Storage
}

func NewSimulator(strg storage.Storage) *Simulator {
	return &Simulator{
		Mutex:    sync.Mutex{},
		spreader: NewSpreader(),
		storage:  strg,
	}
}

func (s *Simulator) Get(key *api.Key) (*api.Message, error) {
	s.Lock()
	defer s.Unlock()

	return s.storage.Get(key)
}

func (s *Simulator) GetAll(key *api.Key) ([]*api.Message, error) {
	s.Lock()
	defer s.Unlock()

	return s.storage.GetAll(key)
}

func (s *Simulator) Put(mes *api.Message) error {
	s.Lock()
	defer s.Unlock()

	if err := s.storage.Put(mes); err != nil {
		return err
	}
	s.spreader.Spread(mes)

	return nil
}

func (s *Simulator) Delete(key *api.Key) error {
	s.Lock()
	defer s.Unlock()

	return s.storage.Delete(key)
}

func (s *Simulator) Watch(key *api.Key, ch *Channel) error {
	s.Lock()
	defer s.Unlock()

	if err := s.spreader.AddWatcher(key, ch); err != nil {
		return err
	}
	return nil
}
