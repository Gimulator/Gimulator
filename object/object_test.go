package object

import (
	"fmt"
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

func TestKeyStringer(t *testing.T) {

	var tests = []struct {
		key  Key
		want string
	}{
		{KeyComplete, "{Type: t, Namespace: ns, Name: n}"},
		{KeyEmpty, "{Type: , Namespace: , Name: }"},
		{KeyOnlyType, "{Type: t, Namespace: , Name: }"},
		{KeyOnlyNamespace, "{Type: , Namespace: ns, Name: }"},
		{KeyOnlyName, "{Type: , Namespace: , Name: n}"},
		{KeyTypeNamespace, "{Type: t, Namespace: ns, Name: }"},
		{KeyTypeName, "{Type: t, Namespace: , Name: n}"},
		{KeyNamespaceName, "{Type: , Namespace: ns, Name: n}"},
	}

	t.Log("Given the need to test method String of Key type.")

	for _, test := range tests {
		t.Logf("\tWhen checking \"%v\"", test.key)

		got := test.key.String()

		if got != test.want {
			t.Errorf(LogFailed(got, test.want, ballotX))
		}else{
			t.Logf(LogApproved(test.want, checkMark))	
		}
	}
}

func TestEqual(t *testing.T) {

	var tests = []struct {
		key1 Key
		key2 *Key
		want bool
	}{
		{KeyEmpty, &Key{}, true},
		{KeyOnlyType, &Key{Type: "t"}, true},
		{KeyComplete, &Key{"t", "ns", "n"}, true},
		{KeyComplete, &Key{"T", "Ns", "N"}, false},
		{KeyOnlyName, &Key{"t", "Namespace", "Name"}, false},
	}

	t.Log("Given the need to test method Equal of Key type.")

	for _, test := range tests {
		t.Logf("\tWhen comparing \"%v\" to \"%v\"", test.key1, test.key2)

		got := test.key1.Equal(test.key2)

		if got != test.want {
			t.Errorf(LogFailed(got, test.want, ballotX))
		}else{
			t.Logf(LogApproved(test.want, checkMark))	
		}
	}
}

func TestMatch(t *testing.T) {

	var tests = []struct {
		key1 Key
		key2 *Key
		want bool
	}{
		{KeyComplete, &Key{"t", "ns", "n"}, true},
		{KeyOnlyType, &Key{Type: "t"}, true},
		{KeyOnlyType, &Key{"t", "a", "b"}, true},
		{KeyEmpty, &Key{}, true},
		{KeyEmpty, &Key{"a", "b", "c"}, true},
		{KeyOnlyType, &KeyTypeNamespace, true},
		{KeyTypeNamespace, &KeyOnlyName, false},
		{KeyComplete, &Key{"T", "Ns", "N"}, false},
		{KeyComplete, &KeyOnlyType, false},
	}	

	t.Log("Given the needed to test mehod Match of Key type.")

	for _, test := range tests {
		t.Logf("\tWhen comparing \"%v\" to \"%v\"", test.key1, test.key2)

		got := test.key1.Match(test.key2)

		if got != test.want {
			t.Errorf(LogFailed(got, test.want, ballotX))
		} else {
			t.Logf(LogApproved(test.want, checkMark))
		}
	}
}

func TestObjectStringer(t *testing.T) {

	var tests = []struct {
		o    Object
		want string
	}{
		{Object{Key: &KeyOnlyType, Value: "Value"}, "{Key: {Type: t, Namespace: , Name: }, Value: 'Value'}"},
		{Object{Key: &KeyComplete, Value: "thisisatestvalue"}, "{Key: {Type: t, Namespace: ns, Name: n}, Value: 'thisisates...'}"},
		{Object{}, "{Key: <nil>, Value: ''}"},
		{Object{Key: &KeyOnlyName, Value: ""}, "{Key: {Type: , Namespace: , Name: n}, Value: ''}"},
	}

	t.Log("Given the needed to test method String of Object type.")

	for _, test := range tests {
		t.Logf("\tWhen checking \"%v\"", test.o)

		got := test.o.String()

		if got != test.want {
			t.Errorf(LogFailed(got, test.want, ballotX))
		}else{
			t.Logf(LogApproved(test.want, checkMark))
		}
	}
}

var (
	KeyComplete = Key{"t", "ns", "n"}
	KeyEmpty = Key{}
	KeyOnlyType = Key{Type: "t"}
	KeyOnlyNamespace = Key{Namespace : "ns"}
	KeyOnlyName = Key{Name : "n"}
	KeyTypeNamespace = Key{Type : "t", Namespace : "ns"}
	KeyTypeName = Key{Type : "t", Name : "n"}
	KeyNamespaceName = Key{Namespace : "ns", Name : "n"}
)
