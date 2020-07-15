package simulator

import (
	"fmt"
	"github.com/Gimulator/Gimulator/object"
	"reflect"
	"testing"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

func LogApproved(want interface{}, checkMark string) string {
	return fmt.Sprintf("\t\tShould have a \"%v\" %v", want, checkMark)
}

func LogFailed(got, want interface{}, ballotX string) string {
	return fmt.Sprintf("\t\tgot %v ***** want %v %v", got, want, ballotX)
}

func TestNewWatcher(t *testing.T) {
	tempCh := make(chan *object.Object)
	var tests = []struct {
		ch          chan *object.Object
		wantWatcher watcher
		wantErr     error
	}{
		{nil, watcher{}, fmt.Errorf("nil channel for creating new watcher")},
		{tempCh, makeTestWatcher(tempCh), nil},
	}
	t.Logf("Given the need to test newWatcher method of watcher type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v\"", test.ch)
		gotWatcher, gotErr := newWatcher(test.ch)

		if !reflect.DeepEqual(gotErr, test.wantErr) || !reflect.DeepEqual(gotWatcher, test.wantWatcher) {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		} else {
			t.Logf(LogApproved(test.wantWatcher, checkMark))
		}
	}
}

func TestSendIfNeeded(t *testing.T) {
	t.Logf("Given the need to test sendIfNeeded method of watcher type.")

	t.Run("Call sendIfNeeded() only one time", func(t *testing.T) {

		var tests = []struct {
			obj  *object.Object
			want error
		}{
			{&ObjectKEmpty, nil},
			{&ObjectKOnlyName, nil},
			{&ObjectKComplete, nil},
			{&ObjectKNamespaceName, nil},
		}

		for _, test := range tests {
			t.Logf("\tWhen checking the value \"%v\"", test.obj)
			w := makeTestWatcher(make(chan *object.Object, 1), &KeyNamespaceName, &KeyComplete)
			got := w.sendIfNeeded(test.obj)

			go func() {
				<-w.ch
			}()

			if reflect.DeepEqual(got, test.want) {
				t.Logf(LogApproved(test.want, checkMark))
			} else {
				t.Errorf(LogFailed(got, test.want, ballotX))
			}
		}
	})

	t.Run("Call sendIfNeeded() more than one time", func(t *testing.T) {

		var tests = []struct {
			obj  *object.Object
			want error
		}{
			{&ObjectKEmpty, nil},
			{&ObjectKOnlyName, nil},
			{&ObjectKComplete, fmt.Errorf("could not write to object")},
		}

		for _, test := range tests {
			t.Logf("\tWhen checking the value \"%v\"", test.obj)
			w := makeTestWatcher(make(chan *object.Object, 1), &KeyComplete)

			got := w.sendIfNeeded(test.obj)
			got = w.sendIfNeeded(test.obj)
			go func() {
				<-w.ch
			}()

			if reflect.DeepEqual(got, test.want) {
				t.Logf(LogApproved(test.want, checkMark))
			} else {
				t.Errorf(LogFailed(got, test.want, ballotX))
			}
		}

	})

}

func TestAddWatch(t *testing.T) {
	w := makeTestWatcher(make(chan *object.Object), &KeyComplete)
	var tests = []struct {
		key *object.Key
	}{
		{&KeyComplete},
		{&KeyOnlyType},
	}
	t.Logf("Given the need to test addWatch method of watcher type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v\"", test.key)
		w.addWatch(test.key)

		b := false
		for _, k := range w.keys {
			if reflect.DeepEqual(k, test.key) {
				t.Logf(LogApproved(test.key, checkMark))
				b = true
				break
			}
		}
		if !b {
			t.Errorf(LogFailed("Key is not in the watcher value!", test.key, checkMark))
		}
	}

}

func makeTestWatcher(ch chan *object.Object, keys ...*object.Key) watcher {
	w := watcher{
		keys: make([]*object.Key, 0),
		ch:   ch,
	}
	for _, k := range keys {
		w.keys = append(w.keys, k)
	}
	return w
}

var (
	KeyComplete          = object.Key{"t", "ns", "n"}
	KeyComplete2         = object.Key{"t2", "ns2", "n2"}
	KeyComplete3         = object.Key{"t3", "ns3", "n3"}
	KeyEmpty             = object.Key{}
	KeyOnlyType          = object.Key{Type: "t"}
	KeyOnlyNamespace     = object.Key{Namespace: "ns"}
	KeyOnlyName          = object.Key{Name: "n"}
	KeyTypeNamespace     = object.Key{Type: "t", Namespace: "ns"}
	KeyTypeName          = object.Key{Type: "t", Name: "n"}
	KeyNamespaceName     = object.Key{Namespace: "ns", Name: "n"}
	ObjectKComplete      = object.Object{Key: &KeyComplete}
	ObjectKComplete2     = object.Object{Key: &KeyComplete2}
	ObjectKComplete3     = object.Object{Key: &KeyComplete3}
	ObjectKEmpty         = object.Object{Key: &KeyEmpty}
	ObjectKOnlyType      = object.Object{Key: &KeyOnlyType}
	ObjectKOnlyNamespace = object.Object{Key: &KeyOnlyNamespace}
	ObjectKOnlyName      = object.Object{Key: &KeyOnlyName}
	ObjectKTypeNamespace = object.Object{Key: &KeyTypeNamespace}
	ObjectKTypeName      = object.Object{Key: &KeyTypeName}
	ObjectKNamespaceName = object.Object{Key: &KeyNamespaceName}
)
