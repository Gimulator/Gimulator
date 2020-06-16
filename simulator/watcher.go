package simulator

import (
	"fmt"

	"github.com/Gimulator/Gimulator/object"
)

type watcher struct {
	keys []*object.Key
	ch   chan *object.Object
}

func newWatcher(ch chan *object.Object) (watcher, error) {
	if ch == nil {
		return watcher{}, fmt.Errorf("nil channel for creating new watcher")
	}

	return watcher{
		keys: make([]*object.Key, 0),
		ch:   ch,
	}, nil
}

func (w *watcher) sendIfNeeded(obj *object.Object) error {
	for _, k := range w.keys {
		if !k.Match(obj.Key) {
			continue
		}

		select {
		case w.ch <- obj:
		default:
			return fmt.Errorf("could not write to object")
		}
		break
	}
	return nil
}

func (w *watcher) addWatch(key *object.Key) {
	for _, k := range w.keys {
		if k.Equal(key) {
			return
		}
	}
	w.keys = append(w.keys, key)
}
