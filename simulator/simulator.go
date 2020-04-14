package simulator

import (
	"sync"

	"gitlab.com/Syfract/Xerac/gimulator/object"
	"gitlab.com/Syfract/Xerac/gimulator/storage"
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

func (s *Simulator) Get(key object.Key) (object.Object, error) {
	s.Lock()
	s.Unlock()

	return s.storage.Get(key)
}

func (s *Simulator) Set(obj object.Object) error {
	s.Lock()
	s.Unlock()

	err := s.storage.Set(obj)
	if err != nil {
		return err
	}
	s.spreader.Spread(obj)
	return nil
}

func (s *Simulator) Delete(key object.Key) error {
	s.Lock()
	s.Unlock()

	return s.storage.Delete(key)
}

func (s *Simulator) Find(key object.Key) ([]object.Object, error) {
	s.Lock()
	s.Unlock()

	return s.storage.Find(key)
}

func (s *Simulator) Watch(key object.Key, ch chan object.Object) {
	s.Lock()
	s.Unlock()

	s.spreader.AddWatcher(key, ch)
}
