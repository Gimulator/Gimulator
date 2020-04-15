package simulator

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/Syfract/Xerac/gimulator/object"
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
		log:      logrus.WithField("Entity", "spreader"),
	}
}

func (s *spreader) AddWatcher(key *object.Key, ch chan *object.Object) {
	s.log.Info("Add new watcher")
	watcher := watcher{
		key: key,
		ch:  ch,
	}

	s.watchers = append(s.watchers, watcher)
}

func (s *spreader) Spread(obj *object.Object) {
	s.log.Info("Start to spread")
	key := *obj.Key
	for _, w := range s.watchers {
		if w.key.Match(&key) {
			select {
			case w.ch <- obj:
			default:
				s.log.Debug("can not write to channel")
			}
		}
	}
}
