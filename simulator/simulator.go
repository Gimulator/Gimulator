package simulator

import (
	"sync"
	"time"

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
		spreader: NewSpreader(),
		storage:  strg,
	}
}

func (s *Simulator) Get(id string, key *object.Key) (*object.Object, error) {
	s.Lock()
	defer s.Unlock()

	return s.storage.Get(key)
}

func (s *Simulator) Set(id string, obj *object.Object) error {
	s.Lock()
	defer s.Unlock()

	obj.Meta = &object.Meta{
		Owner:        id,
		CreationTime: time.Now(),
		Method:       object.MethodSet,
	}

	if err := s.storage.Set(obj); err != nil {
		return err
	}
	s.spreader.Spread(obj)

	return nil
}

func (s *Simulator) Delete(id string, key *object.Key) error {
	s.Lock()
	defer s.Unlock()

	return s.storage.Delete(key)
}

func (s *Simulator) Find(id string, key *object.Key) ([]*object.Object, error) {
	s.Lock()
	defer s.Unlock()

	return s.storage.Find(key)
}

func (s *Simulator) Watch(id string, key *object.Key, ch chan *object.Object) error {
	s.Lock()
	defer s.Unlock()

	if err := s.spreader.AddWatcher(id, key, ch); err != nil {
		return err
	}
	return nil
}
