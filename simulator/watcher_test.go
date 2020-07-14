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
		{tempCh, watcher{make([]*object.Key, 0), tempCh}, nil},
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

// INCOMPLETE
/*
func TestSendIfNeeded(t *testing.T) {

	var tests = []struct {
		obj     *object.Object
		want error
	}{
		{&ObjectKEmpty, nil},
		{&ObjectKOnlyName, nil},
		{&ObjectKComplete, nil},
		{&ObjectKNamespaceName, nil},
	}

	t.Logf("Given the need to test sendIfNeeded method of watcher type.")

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v\"", test.obj)

		ch := make(chan *object.Object)
		w, _ := newWatcher(ch)
		w.keys = []*object.Key{
			&KeyNamespaceName,
			&KeyComplete,
		}

		var(
			got error
			wg sync.WaitGroup
		)

		wg.Add(1)
		go func(){
			got = w.sendIfNeeded(test.obj)
			defer wg.Done()
		}()
		go func(){
			<- w.ch
		}()
		wg.Wait()

		if reflect.DeepEqual(got, test.want){
			t.Logf(LogApproved(test.want, checkMark))
		} else {
			t.Errorf(LogFailed(got, test.want, ballotX))
		}

		close(ch)
	}

}
*/

var (
	KeyComplete          = object.Key{"t", "ns", "n"}
	KeyEmpty             = object.Key{}
	KeyOnlyType          = object.Key{Type: "t"}
	KeyOnlyNamespace     = object.Key{Namespace: "ns"}
	KeyOnlyName          = object.Key{Name: "n"}
	KeyTypeNamespace     = object.Key{Type: "t", Namespace: "ns"}
	KeyTypeName          = object.Key{Type: "t", Name: "n"}
	KeyNamespaceName     = object.Key{Namespace: "ns", Name: "n"}
	ObjectKComplete      = object.Object{Key: &KeyComplete}
	ObjectKEmpty         = object.Object{Key: &KeyEmpty}
	ObjectKOnlyType      = object.Object{Key: &KeyOnlyType}
	ObjectKOnlyNamespace = object.Object{Key: &KeyOnlyNamespace}
	ObjectKOnlyName      = object.Object{Key: &KeyOnlyName}
	ObjectKTypeNamespace = object.Object{Key: &KeyTypeNamespace}
	ObjectKTypeName      = object.Object{Key: &KeyTypeName}
	ObjectKNamespaceName = object.Object{Key: &KeyNamespaceName}
)
