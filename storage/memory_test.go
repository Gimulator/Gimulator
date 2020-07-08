package storage

import (
	"fmt"
	"github.com/Gimulator/Gimulator/object"
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

func TestValidateKey(t *testing.T) {

	m := NewMemory()

	var tests = []struct {
		key  *object.Key
		want error
	}{
		{&KeyComplete, nil},
		{&KeyEmpty, fmt.Errorf("invalid key with empty Name")},
		{&KeyOnlyType, fmt.Errorf("invalid key with empty Name")},
		{&KeyOnlyNamespace, fmt.Errorf("invalid key with empty Name")},
		{&KeyOnlyName, fmt.Errorf("invalid key with empty Namespace")},
		{&KeyTypeNamespace, fmt.Errorf("invalid key with empty Name")},
		{&KeyTypeName, fmt.Errorf("invalid key with empty Namespace")},
		{&KeyNamespaceName, fmt.Errorf("invalid key with empty Type")},
	}

	t.Logf("Given the need to test method validateKey of Memory type.")

	for _, test := range tests {
		t.Logf("\tWhen checking \"%v\"", test.key)

		got := m.validateKey(test.key)

		if got == nil && test.want != nil {
			t.Errorf(LogFailed(got, test.want, ballotX))
		} else if got != nil && test.want == nil {
			t.Errorf(LogFailed(got, test.want, ballotX))
		} else if got == nil && test.want == nil {
			t.Logf(LogApproved(test.want, checkMark))
			continue
		} else if got.Error() != test.want.Error() {
			t.Errorf(LogFailed(got, test.want, ballotX))
		} else {
			t.Logf(LogApproved(test.want, checkMark))
		}
	}
}

func TestGet(t *testing.T) {

	m := Memory{
		storage: make(map[object.Key]*object.Object),
	}
	m.storage[KeyComplete] = &ObjectKComplete

	var tests = []struct {
		key     *object.Key
		wantObj *object.Object
		wantErr error
	}{
		{&KeyComplete, &ObjectKComplete, nil},
		{&KeyEmpty, nil, fmt.Errorf("invalid key with empty Name")},
		{&KeyOnlyType, nil, fmt.Errorf("invalid key with empty Name")},
		{&object.Key{"type", "namespace", "name"}, nil, fmt.Errorf("object with key={Type: type, Namespace: namespace, Name: name} does not exist")},
	}

	t.Log("Given the need to test get method of Memory type.")

	for _, test := range tests {
		t.Logf("\tWhen checking \"%v\"", test.key)

		gotObj, gotErr := m.get(test.key)

		if gotErr != nil {
			if test.wantErr == nil {
				t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
			} else if test.wantErr.Error() == gotErr.Error() {
				t.Logf(LogApproved(test.wantErr, checkMark))
			} else {
				t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
			}
		} else if gotObj != test.wantObj {
			t.Errorf(LogFailed(gotObj, test.wantObj, ballotX))
		} else {
			t.Logf(LogApproved(test.wantObj, checkMark))
		}
	}
}

func TestSet(t *testing.T) {
	m := Memory{
		storage: make(map[object.Key]*object.Object),
	}

	m.storage[KeyComplete] = &ObjectKComplete

	var tests = []struct {
		obj  *object.Object
		want error
	}{
		{&ObjectKComplete, nil},
		{&ObjectKEmpty, fmt.Errorf("invalid key with empty Name")},
		{&ObjectKOnlyType, fmt.Errorf("invalid key with empty Name")},
		{&object.Object{Key: &object.Key{"type", "namespace", "name"}, Value: "val"}, nil},
	}

	t.Log("Given the need to test set method of Memory type.")

	for _, test := range tests {
		t.Logf("\tWhen checking \"%v\"", test.obj)

		got := m.set(test.obj)

		if got != nil {
			if got.Error() == test.want.Error() {
				t.Logf(LogApproved(test.want, checkMark))
			} else {
				t.Errorf(LogFailed(got, test.want, ballotX))
			}
		} else if got := m.storage[*test.obj.Key]; got == test.obj {
			t.Logf(LogApproved(test.want, checkMark))
		} else {
			t.Errorf(LogFailed("saved item in memory", test.obj, ballotX))
		}
	}

}

func TestDelete(t *testing.T) {
	m := Memory{
		storage: make(map[object.Key]*object.Object),
	}

	m.storage[KeyComplete] = &ObjectKComplete

	var tests = []struct {
		key  *object.Key
		want error
	}{
		{&KeyComplete, nil},
		{&object.Key{"type", "namespace", "name"}, fmt.Errorf("object with key={Type: type, Namespace: namespace, Name: name} does not exist")},
		{&KeyEmpty, fmt.Errorf("invalid key with empty Name")},
		{&KeyOnlyType, fmt.Errorf("invalid key with empty Name")},
	}

	t.Log("Given the need to test delete method of Memory type.")

	for _, test := range tests {
		t.Logf("\tWhen checking \"%v\"", test.key)

		got := m.delete(test.key)

		if got != nil {
			if got.Error() == test.want.Error() {
				t.Logf(LogApproved(test.want, checkMark))
			} else {
				t.Errorf(LogFailed(got, test.want, ballotX))
			}
		} else if got := m.storage[*test.key]; got == nil {
			t.Logf(LogApproved(test.want, checkMark))
		} else {
			t.Errorf(LogFailed("nothing in memory", test.key, ballotX))
		}
	}
}

func TestFind(t *testing.T) {
	m := Memory{
		storage: make(map[object.Key]*object.Object),
	}

	m.storage[KeyOnlyName] = &ObjectKOnlyName
	m.storage[KeyOnlyType] = &ObjectKOnlyType
	m.storage[KeyTypeNamespace] = &ObjectKTypeNamespace

	var tests = []struct {
		key  *object.Key
		want []*object.Object
	}{
		{&KeyOnlyName, []*object.Object{&ObjectKOnlyName}},
		{&KeyOnlyType, []*object.Object{&ObjectKOnlyType, &ObjectKTypeNamespace}},
		{&KeyTypeNamespace, []*object.Object{&ObjectKTypeNamespace}},
		{&KeyComplete, make([]*object.Object, 0)},
	}

	t.Log("Given the need to test find method of Memory type.")

	for _, test := range tests {
		t.Logf("\tWhen checking \"%v\"", test.key)

		got := m.find(test.key)

		if len(got) == 0 && len(test.want) == 0 {
			t.Logf(LogApproved(test.want, checkMark))
		} else {
			var flag = false

			for _, vGot := range got {
				flag = false
				for _, vTest := range test.want {
					if vGot == vTest {
						flag = true
						break
					}
				}
				if flag == false {
					t.Fatalf(LogFailed(got, test.want, ballotX))
				}
			}

			if flag == false {
				t.Errorf(LogFailed(got, test.want, ballotX))
			} else {
				t.Logf(LogApproved(test.want, checkMark))
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
	ObjectKComplete      = object.Object{Key: &KeyComplete}
	ObjectKEmpty         = object.Object{Key: &KeyEmpty}
	ObjectKOnlyType      = object.Object{Key: &KeyOnlyType}
	ObjectKOnlyNamespace = object.Object{Key: &KeyOnlyNamespace}
	ObjectKOnlyName      = object.Object{Key: &KeyOnlyName}
	ObjectKTypeNamespace = object.Object{Key: &KeyTypeNamespace}
	ObjectKTypeName      = object.Object{Key: &KeyTypeName}
	ObjectKNamespaceName = object.Object{Key: &KeyNamespaceName}
)
