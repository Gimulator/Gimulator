package simulator

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/Gimulator/Gimulator/object"
	"github.com/Gimulator/Gimulator/storage"
	"reflect"
	"sync"
	"testing"
)

func TestNewSimulator(t *testing.T) {
	strg := makeTestStorage()
	want := makeTestSimulator(strg)
	got := NewSimulator(strg)

	if reflect.DeepEqual(got, want) {
		t.Logf(LogApproved(want, checkMark))
	} else {
		t.Errorf(LogFailed(got, want, ballotX))
	}
}

func TestGet(t *testing.T) {
	s := makeTestSimulator(makeTestStorage(&ObjectKComplete))
	tests := []struct {
		id      string
		key     *object.Key
		wantObj *object.Object
		wantErr error
	}{
		{id, &KeyComplete, &ObjectKComplete, nil},
		{id1, &KeyEmpty, nil, fmt.Errorf("error")},
		{id2, &KeyComplete2, nil, fmt.Errorf("error")},
	}
	t.Logf("Given the need to test Get method of Simulator type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v, %v\"", test.key, test.id)
		gotObj, gotErr := s.Get(test.id, test.key)

		if reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) && reflect.DeepEqual(gotObj, test.wantObj) {
			t.Logf(LogApproved(test.wantObj, checkMark))
		} else if !reflect.DeepEqual(gotErr, test.wantErr) {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		} else {
			t.Errorf(LogFailed(gotObj, test.wantObj, ballotX))
		}
	}

}

func TestSet(t *testing.T) {
	s := makeTestSimulator(makeTestStorage(&ObjectKComplete))
	tests := []struct {
		id      string
		obj     *object.Object
		key     *object.Key
		wantObj *object.Object
		wantErr error
	}{
		{id1, &ObjectKComplete, &KeyComplete, &ObjectKComplete, nil},
		{id2, &ObjectKEmpty, &KeyEmpty, nil, fmt.Errorf("error")},
	}
	t.Logf("Given the need to test Set method of Simulator type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v, %v\"", test.id, test.key)
		gotErr := s.Set(test.id, test.obj)
		gotObj, _ := s.storage.Get(test.key)

		if reflect.DeepEqual(gotObj, test.wantObj) && reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) {
			t.Logf(LogApproved(test.wantObj, checkMark))
		} else if reflect.TypeOf(gotErr) != reflect.TypeOf(test.wantErr) {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		} else {
			t.Errorf(LogFailed(gotObj, test.wantObj, ballotX))
		}
	}

}

func TestDelete(t *testing.T) {
	s := makeTestSimulator(makeTestStorage(&ObjectKComplete))
	tests := []struct {
		id      string
		key     *object.Key
		wantObj *object.Object
		wantErr error
	}{
		{id, &KeyComplete2, nil, fmt.Errorf("error")},
		{id1, &KeyComplete, nil, nil},
		{id2, &KeyEmpty, nil, fmt.Errorf("error")},
	}
	t.Logf("Given the need to test Delete method of Simulator type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v, %v\"", test.id, test.key)

		gotErr := s.Delete(test.id, test.key)
		gotObj, _ := s.storage.Get(test.key)

		if reflect.DeepEqual(gotObj, test.wantObj) && reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) {
			t.Logf(LogApproved(test.wantErr, checkMark))
		} else if reflect.TypeOf(gotErr) != reflect.TypeOf(test.wantErr) {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		} else {
			t.Errorf(LogFailed(gotObj, test.wantObj, ballotX))
		}
	}
}

func TestFind(t *testing.T) {
	s := makeTestSimulator(makeTestStorage(&ObjectKComplete, &ObjectKComplete2))
	tests := []struct {
		id      string
		key     *object.Key
		wantObj []*object.Object
		wantErr error
	}{
		{id, &KeyComplete3, make([]*object.Object, 0), nil},
		{id1, &KeyComplete, []*object.Object{&ObjectKComplete}, nil},
		{id2, &KeyEmpty, []*object.Object{&ObjectKComplete2, &ObjectKComplete}, nil},
	}
	t.Logf("Given the need to test Find method of Simulator type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v, %v\"", test.id, test.key)
		gotObj, gotErr := s.Find(test.id, test.key)

		flag := false
		for _, v := range gotObj{
			flag = true
			for _,v2 := range test.wantObj{
				if reflect.DeepEqual(v, v2){
					flag = false
					break
				}
			}
			if flag {
				t.Errorf(LogFailed(gotObj, test.wantObj, ballotX))	
				break
			}
		}
		if reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) {
			t.Logf(LogApproved(test.wantErr, checkMark))
		} else  {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		}
	}
}

func TestWatch(t *testing.T) {
	s := makeTestSimulator(makeTestStorage(&ObjectKComplete))
	s.spreader.watchers[id] = watcher{
		keys: []*object.Key{&KeyComplete},
		ch:   make(chan *object.Object),
	}
	tempch := make(chan *object.Object)
	tests := []struct {
		id          string
		key         *object.Key
		ch          chan *object.Object
		wantWatcher watcher
		wantErr     error
	}{
		{id, &KeyOnlyType, nil, s.spreader.watchers[id], nil},
		{id1, &KeyComplete, tempch, watcher{[]*object.Key{&KeyComplete}, tempch}, nil},
		{id2, &KeyComplete, nil, watcher{}, fmt.Errorf("error")},
	}
	t.Logf("Given the need to test Watch method of Simulator type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v, %v\"", test.id, test.key)
		gotErr := s.Watch(test.id, test.key, test.ch)
		gotWatcher := s.spreader.watchers[test.id]

		if reflect.DeepEqual(gotWatcher, test.wantWatcher) && reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) {
			t.Logf(LogApproved(test.wantErr, checkMark))
		} else if reflect.TypeOf(gotErr) != reflect.TypeOf(test.wantErr) {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		} else {
			t.Errorf(LogFailed(gotWatcher, test.wantWatcher, ballotX))
		}
	}
}

func makeTestSimulator(strg *storage.Memory) *Simulator{
	sp := &spreader{
		watchers: make(map[string]watcher),
		log:      logrus.WithField("entity", "spreader"),
	}
	return &Simulator{
		Mutex:    sync.Mutex{},
		spreader: sp,
		storage:  strg,
	}
}

func makeTestStorage(objs ...*object.Object) *storage.Memory {
	strg := storage.NewMemory()
	for _, obj := range objs {
		strg.Set(obj)
	}
	return strg
}

var (
	id = "id"
	id1 = "id1"
	id2 = "id2"
)