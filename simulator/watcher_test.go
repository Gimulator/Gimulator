package simulator

import (
	"testing"
	"github.com/Gimulator/Gimulator/object"
	"fmt"
	"reflect"
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
		ch chan *object.Object
		wantWatcher watcher
		wantErr error
	}{
		{nil, watcher{}, fmt.Errorf("nil channel for creating new watcher")},
		{tempCh, watcher{make([]*object.Key, 0), tempCh}, nil},
	}

	t.Logf("Given the need to test newWatcher method of watcher type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v\"", test.ch)

		gotWatcher, gotErr := newWatcher(test.ch)

		if gotErr != nil {
			if test.wantErr != nil && gotErr.Error() == test.wantErr.Error() {
				t.Logf(LogApproved(test.wantErr, checkMark))
			} else {
				t.Errorf(LogFailed(gotErr, test.wantErr,ballotX))
			}
		} else {
			if test.wantWatcher.ch == gotWatcher.ch && reflect.ValueOf(gotWatcher.keys).Kind() == reflect.ValueOf(test.wantWatcher.keys).Kind() {
				t.Logf(LogApproved(test.wantWatcher, checkMark))
			} else {
				t.Errorf(LogFailed(gotWatcher, test.wantWatcher, ballotX))
			}
		}
	}
}


var (
	KeyComplete          = object.Key{"t", "ns", "n"}
	KeyEmpty             = object.Key{}
	KeyOnlyType          = object.Key{Type: "t"}
	KeyOnlyNamespace     = object.Key{Namespace: "ns"}
	KeyOnlyName          = object.Key{Name: "n"}
	KeyTypeNamespace     = object.Key{Type: "t", Namespace: "ns"}
	KeyTypeName          = object.Key{Type: "t", Name: "n"}
	KeyNamespaceName     = object.Key{Namespace: "ns", Name: "n"}
)