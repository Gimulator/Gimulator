package simulator

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"gitlab.com/Syfract/Xerac/gimulator/object"
	"gitlab.com/Syfract/Xerac/gimulator/storage"
)

type Simulator struct {
	sync.Mutex
	spreader *spreader
	storage  storage.Storage
	log      *logrus.Entry
}

func NewSimulator(strg storage.Storage) *Simulator {
	return &Simulator{
		Mutex:    sync.Mutex{},
		spreader: Newspreader(),
		storage:  strg,
		log:      logrus.WithField("Entity", "simulator"),
	}
}

func (s *Simulator) Get(key object.Key) (object.Object, error) {
	s.log.Info("Start to handle get")
	s.Lock()
	s.Unlock()

	return s.storage.Get(key)
}

func (s *Simulator) Set(obj object.Object) error {
	s.log.Info("Start to handle set")
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
	s.log.Info("Start to handle delete")
	s.Lock()
	s.Unlock()

	return s.storage.Delete(key)
}

func (s *Simulator) Find(key object.Key) ([]object.Object, error) {
	s.log.Info("Start to handle find")
	s.Lock()
	s.Unlock()

	return s.storage.Find(key)
}

func (s *Simulator) Watch(key object.Key, ch chan object.Object) error {
	s.log.Info("Start to handle watch")
	s.Lock()
	s.Unlock()

	if ch == nil {
		s.log.Error("nil channel for watch command")
		return fmt.Errorf("nil channel")
	}

	s.spreader.AddWatcher(key, ch)
	return nil
}
