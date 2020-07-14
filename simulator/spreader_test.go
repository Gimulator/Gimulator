package simulator

import (
	"fmt"
	"github.com/Gimulator/Gimulator/object"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestNewSpreader(t *testing.T) {
	want := &spreader{
		watchers: make(map[string]watcher),
		log:      logrus.WithField("entity", "spreader"),
	}

	t.Logf("Given the need to test NewSpreader function of watcher type.")

	got := NewSpreader()

	if reflect.DeepEqual(got, want) {
		t.Logf(LogApproved(want, checkMark))
	} else {
		t.Errorf(LogFailed(got, want, ballotX))
	}
}

func TestAddWatcher(t *testing.T) {
	s := &spreader{
		watchers: make(map[string]watcher),
		log:      logrus.WithField("entity", "spreader"),
	}
	s.watchers["test"] = watcher{
		keys: []*object.Key{&KeyComplete, &KeyOnlyType},
		ch:   make(chan *object.Object),
	}

	testCh := make(chan *object.Object)
	var tests = []struct {
		id          string
		key         *object.Key
		ch          chan *object.Object
		wantErr     error
		wantWatcher watcher
	}{
		{"test", &KeyOnlyName, nil, nil, s.watchers["test"]},
		{"test", &KeyOnlyName, make(chan *object.Object), nil, s.watchers["test"]},
		{"test1", &KeyOnlyType, nil, fmt.Errorf("nil channel for creating new watcher"), watcher{}},
		{"test2", &KeyOnlyType, testCh, nil, watcher{[]*object.Key{&KeyOnlyType}, testCh}},
	}

	t.Logf("Given the need to test addWatcher method of spreader type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v, %v, %v\"", test.id, test.key, test.ch)

		gotErr := s.AddWatcher(test.id, test.key, test.ch)
		gotWatcher := s.watchers[test.id]

		if reflect.DeepEqual(gotErr, test.wantErr) && reflect.DeepEqual(gotWatcher, test.wantWatcher) {
			t.Logf(LogApproved(test.wantWatcher, checkMark))
		} else if !reflect.DeepEqual(gotErr, test.wantErr) {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		} else {
			t.Errorf(LogFailed(gotWatcher, test.wantWatcher, ballotX))
		}
	}
}
