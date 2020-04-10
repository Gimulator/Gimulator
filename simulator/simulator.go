package simulator

import (
	"gitlab.com/Syfract/Xerac/gimulator/object"
	"gitlab.com/Syfract/Xerac/gimulator/spread"
	"gitlab.com/Syfract/Xerac/gimulator/storage"
)

type Simulator struct {
	spreader *spread.Spreader
	storage  storage.Storage
}

func NewSimulator() *Simulator {
	return &Simulator{
		spreader: spread.NewSpreader(),
		storage:  storage.NewMemory(),
	}
}

func (s *Simulator) Get(key object.Key) (object.Object, error) {
	return s.storage.Get(key)
}

func (s *Simulator) Set(obj object.Object) error {
	err := s.storage.Set(obj)
	if err != nil {
		return err
	}
	s.spreader.Spread(obj)
	return nil
}

func (s *Simulator) Delete(key object.Key) error {
	return s.storage.Delete(key)
}

func (s *Simulator) Find(key object.Key) ([]object.Object, error) {
	return s.storage.Find(key)
}

func (s *Simulator) Watch(key object.Key, ch chan object.Object) {
	s.spreader.AddWatcher(key, ch)
}
