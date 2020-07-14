package simulator

import (
	"fmt"
	"github.com/Gimulator/Gimulator/object"
	"github.com/Gimulator/Gimulator/storage"
	"reflect"
	"sync"
	"testing"
)

func TestNewSimulator(t *testing.T) {
	strg := storage.NewMemory()
	want := &Simulator{
		Mutex:    sync.Mutex{},
		spreader: NewSpreader(),
		storage:  strg,
	}

	got := NewSimulator(strg)

	if reflect.DeepEqual(got, want) {
		t.Logf(LogApproved(want, checkMark))
	} else {
		t.Errorf(LogFailed(got, want, ballotX))
	}
}

func TestGet(t *testing.T) {
	strg := storage.NewMemory()
	strg.Set(&ObjectKComplete)
	s := &Simulator{
		Mutex:    sync.Mutex{},
		spreader: NewSpreader(),
		storage:  strg,
	}

	tests := []struct {
		key     *object.Key
		id      string
		wantObj *object.Object
		wantErr error
	}{
		{&KeyComplete, "test id", &ObjectKComplete, nil},
		{&KeyEmpty, "another id", nil, fmt.Errorf("error")},
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
