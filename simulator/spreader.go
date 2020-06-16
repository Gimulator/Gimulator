package simulator

import (
	"github.com/Gimulator/Gimulator/object"
	"github.com/sirupsen/logrus"
)

type spreader struct {
	watchers map[string]watcher
	log      *logrus.Entry
}

func NewSpreader() *spreader {
	return &spreader{
		watchers: make(map[string]watcher),
		log:      logrus.WithField("entity", "spreader"),
	}
}

func (s *spreader) AddWatcher(id string, key *object.Key, ch chan *object.Object) error {
	if w, exists := s.watchers[id]; exists {
		w.addWatch(key)
		return nil
	}

	w, err := newWatcher(ch)
	if err != nil {
		return err
	}
	w.addWatch(key)
	s.watchers[id] = w

	return nil
}

func (s *spreader) Spread(obj *object.Object) {
	s.log.WithField("object", obj.String()).Debug("starting to write objects to channels")

	for id, w := range s.watchers {
		if err := w.sendIfNeeded(obj); err != nil {
			s.log.WithField("id", id).WithField("object", obj.String()).Error(err)
		}
	}
}
