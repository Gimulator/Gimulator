package simulator

import (
	"gitlab.com/Syfract/Xerac/gimulator/object"
)

type watcher struct {
	key object.Key
	ch  chan object.Object
}

type spreader struct {
	watchers []watcher
}

func Newspreader() *spreader {
	return &spreader{
		watchers: make([]watcher, 0),
	}
}

func (s *spreader) AddWatcher(key object.Key, ch chan object.Object) {
	watcher := watcher{
		key: key,
		ch:  ch,
	}

	s.watchers = append(s.watchers, watcher)
}

func (s *spreader) Spread(obj object.Object) {
	key := obj.Key
	for _, w := range s.watchers {
		if w.key.Match(key) {
			select {
			case w.ch <- obj:
			default:
			}
		}
	}
}
