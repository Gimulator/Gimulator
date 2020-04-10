package spread

import (
	"sync"

	"gitlab.com/Syfract/Xerac/gimulator/object"
)

type watcher struct {
	key object.Key
	ch  chan object.Object
}

type Spreader struct {
	sync.Mutex
	watchers []watcher
}

func NewSpreader() *Spreader {
	return &Spreader{
		Mutex:    sync.Mutex{},
		watchers: make([]watcher, 0),
	}
}

func (s *Spreader) AddWatcher(key object.Key, ch chan object.Object) {
	s.Lock()
	defer s.Unlock()

	watcher := watcher{
		key: key,
		ch:  ch,
	}

	s.watchers = append(s.watchers, watcher)
}

func (s *Spreader) Spread(obj object.Object) {
	s.Lock()
	defer s.Unlock()

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
