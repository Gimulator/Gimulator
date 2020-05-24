package simulator

import (
	"fmt"

	"github.com/Gimulator/Gimulator/object"
	"github.com/sirupsen/logrus"
)

type watcher struct {
	key *object.Key
	ch  chan *object.Object
}

type spreader struct {
	watchers []watcher
	log      *logrus.Entry
}

func Newspreader() *spreader {
	return &spreader{
		watchers: make([]watcher, 0),
		log:      logrus.WithField("entity", "spreader"),
	}
}

func (s *spreader) AddWatcher(key *object.Key, ch chan *object.Object) error {
	if ch == nil {
		return fmt.Errorf("nil channel")
	}

	watcher := watcher{
		key: key,
		ch:  ch,
	}

	s.watchers = append(s.watchers, watcher)
	return nil
}

func (s *spreader) Spread(obj *object.Object) {
	s.log.Debug("starting to write objects to channels")

	key := *obj.Key
	for _, w := range s.watchers {
		if w.key.Match(&key) {
			select {
			case w.ch <- obj:
			default:
				s.log.WithField("object", obj.String()).Error("can not write to channel")
			}
		}
	}
}
