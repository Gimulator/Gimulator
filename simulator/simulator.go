package simulator

import (
	"sync"

	"github.com/Gimulator/Gimulator/object"
	"github.com/Gimulator/Gimulator/storage"
)

type Simulator struct {
	sync.Mutex
	spreader *spreader
	storage  storage.Storage
}

func NewSimulator(strg storage.Storage) *Simulator {
	return &Simulator{
		Mutex:    sync.Mutex{},
		spreader: Newspreader(),
		storage:  strg,
	}
}

func (s *Simulator) Get(key *object.Key) (*object.Object, error) {
	s.Lock()
	defer s.Unlock()

	return s.storage.Get(key)
}

func (s *Simulator) Set(obj *object.Object) error {
	s.Lock()
	defer s.Unlock()

	if err := s.storage.Set(obj); err != nil {
		return err
	}
	s.spreader.Spread(obj)

	return nil
}

func (s *Simulator) Delete(key *object.Key) error {
	s.Lock()
	defer s.Unlock()

	return s.storage.Delete(key)
}

func (s *Simulator) Find(key *object.Key) ([]*object.Object, error) {
	s.Lock()
	defer s.Unlock()

	return s.storage.Find(key)
}

func (s *Simulator) Watch(key *object.Key, ch chan *object.Object) error {
	s.Lock()
	defer s.Unlock()

	if err := s.spreader.AddWatcher(key, ch); err != nil {
		return err
	}
	return nil
}
