package simulator

//import (
//	"fmt"
//	"github.com/Gimulator/Gimulator/object"
//	"github.com/sirupsen/logrus"
//	"reflect"
//	"testing"
//)
//
//func TestNewSpreader(t *testing.T) {
//	want := &spreader{
//		watchers: make(map[string]watcher),
//		log:      logrus.WithField("entity", "spreader"),
//	}
//	t.Logf("Given the need to test NewSpreader function of watcher type.")
//	got := NewSpreader()
//
//	if reflect.DeepEqual(got, want) {
//		t.Logf(LogApproved(want, checkMark))
//	} else {
//		t.Errorf(LogFailed(got, want, ballotX))
//	}
//}
//
//func TestAddWatcher(t *testing.T) {
//	s := &spreader{
//		watchers: make(map[string]watcher),
//		log:      logrus.WithField("entity", "spreader"),
//	}
//	s.watchers[id] = watcher{
//		keys: []*object.Key{&KeyComplete, &KeyOnlyType},
//		channel:   make(chan *object.Object),
//	}
//	testCh := make(chan *object.Object)
//	var tests = []struct {
//		id          string
//		key         *object.Key
//		ch          chan *object.Object
//		wantErr     error
//		wantWatcher watcher
//	}{
//		{id, &KeyOnlyName, nil, nil, s.watchers[id]},
//		{id, &KeyOnlyName, make(chan *object.Object), nil, s.watchers[id]},
//		{id1, &KeyOnlyType, nil, fmt.Errorf("error"), watcher{}},
//		{id2, &KeyOnlyType, testCh, nil, watcher{[]*object.Key{&KeyOnlyType}, testCh}},
//	}
//	t.Logf("Given the need to test addWatcher method of spreader type.")
//
//	for _, test := range tests {
//		t.Logf("\tWhen checking the value \"%v, %v, channel: %v\"", test.id, test.key, test.ch)
//		gotErr := s.AddWatcher(test.id, test.key, test.ch)
//		gotWatcher := s.watchers[test.id]
//
//		if reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) && reflect.DeepEqual(gotWatcher, test.wantWatcher) {
//			t.Logf(LogApproved(test.wantWatcher, checkMark))
//		} else if reflect.TypeOf(gotErr) != reflect.TypeOf(test.wantErr) {
//			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
//		} else {
//			t.Errorf(LogFailed(gotWatcher, test.wantWatcher, ballotX))
//		}
//	}
//}
//
