package auth

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/object"
)

type TestInput struct {
	ConfigPath string
	Pass       bool
	id         string
	hash       string
}

var cp = []string{
	"example/roles.yaml",
	"example/roles1.yaml",
	"example/roles2.yaml",
	"example/roles3.yaml",
	"example/roles4.yaml",
	"example/",
	"WrongPath1",
	"WrongPath2",
}

func GetResult(res bool, e error) (string, bool) {
	if e != nil && res == false {
		return fmt.Sprintf("Test has error and expected to be failed: %s", e), true //true negative
	} else if res == false && e == nil {
		return "Test passed but expected to have error.", false //false positive
	} else if res == true && e != nil {
		return fmt.Sprintf("Test has error but expected to be passed. error: %s", e), false //false negative
	}
	return "Test has no error and passed", true //true positive
}
func PrintResult(t *testing.T, res bool, e error) {
	s, b := GetResult(res, e)
	if b != true {
		t.Errorf("FAILED UNEXPECTEDLY: %s", s)
	} else {
		t.Logf("PASSED: %s", s)
	}
}
func TestNewAuth(t *testing.T) {
	testCases := []TestInput{
		{ConfigPath: cp[0], Pass: true},
		{ConfigPath: cp[1], Pass: true},
		{ConfigPath: cp[2], Pass: true},
		{ConfigPath: cp[5], Pass: false},
		{ConfigPath: cp[6], Pass: false},
		{ConfigPath: cp[7], Pass: false},
	}
	for tc, _ := range testCases {
		for tc, _ := range testCases {
			config, err1 := config.NewConfig(testCases[tc].ConfigPath)
			GetResult(testCases[tc].Pass, err1)
			if err1 != nil {
				PrintResult(t, testCases[tc].Pass, err1)
			} else if _, err2 := auth.NewAuth(config); err2 != nil {
				PrintResult(t, testCases[tc].Pass, err2)
			} else {
				PrintResult(t, testCases[tc].Pass, nil)
			}
		}

	}
}

func TestRegister(t *testing.T) {
	testCases := []TestInput{
		{ConfigPath: cp[0], id: "", Pass: false},
		{ConfigPath: cp[0], id: "agent-123", Pass: true},
		{ConfigPath: cp[0], id: "judge-123", Pass: true},
		{ConfigPath: cp[0], id: "logger-123", Pass: true},
		{ConfigPath: cp[0], id: "abc-123", Pass: false},

		{ConfigPath: cp[1], id: "agent-123", Pass: true},
		{ConfigPath: cp[1], id: "judge-123", Pass: true},
		{ConfigPath: cp[1], id: "logger-123", Pass: true},

		{ConfigPath: cp[2], id: "agent-123", Pass: true},
		{ConfigPath: cp[2], id: "logger-123", Pass: true},
		{ConfigPath: cp[2], id: "judge-123", Pass: true},

		{ConfigPath: cp[2], id: "abc-123", Pass: false},
		{ConfigPath: cp[6], id: "judge-123", Pass: false},
		{ConfigPath: cp[6], id: "logger-123", Pass: false},
		{ConfigPath: cp[7], id: "", Pass: false},
	}

	for tc, _ := range testCases {
		config, err1 := config.NewConfig(testCases[tc].ConfigPath)
		if err1 != nil {
			PrintResult(t, testCases[tc].Pass, err1)
		} else {
			au, err2 := NewAuth(config)
			if err2 != nil {
				PrintResult(t, testCases[tc].Pass, err2)
			} else {
				err3 := au.Register(testCases[tc].id)
				if err3 != nil {
					PrintResult(t, testCases[tc].Pass, err3)
				} else {
					PrintResult(t, testCases[tc].Pass, err2)
				}
			}
		}
	}
}

func TestAuth(t *testing.T) {
	testCases := []TestInput{
		{ConfigPath: cp[0], id: "", Pass: false},
		{ConfigPath: cp[0], id: "agent-123", Pass: true},
		{ConfigPath: cp[0], id: "judge-123", Pass: false},
		{ConfigPath: cp[0], id: "logger-123", Pass: true},
		{ConfigPath: cp[0], id: "abc-123", Pass: false},

		{ConfigPath: cp[1], id: "agent-123", Pass: false},
		{ConfigPath: cp[1], id: "judge-123", Pass: true},
		{ConfigPath: cp[1], id: "logger-123", Pass: true},

		{ConfigPath: cp[2], id: "agent-123", Pass: false},
		{ConfigPath: cp[2], id: "logger-123", Pass: true},
		{ConfigPath: cp[2], id: "judge-123", Pass: true},
		{ConfigPath: cp[2], id: "abc-123", Pass: false},

		{ConfigPath: cp[3], id: "agent-123", Pass: false},
		{ConfigPath: cp[3], id: "logger-123", Pass: false},
		{ConfigPath: cp[3], id: "judge-123", Pass: false},

		{ConfigPath: cp[6], id: "judge-123", Pass: false},
		{ConfigPath: cp[6], id: "logger-123", Pass: false},
		{ConfigPath: cp[7], id: "", Pass: false}, //idk
	}
	for tc, _ := range testCases {
		config, err1 := config.NewConfig(testCases[tc].ConfigPath)
		if err1 != nil {
			PrintResult(t, testCases[tc].Pass, err1)
		} else {
			au, err2 := NewAuth(config)
			if err2 != nil {
				PrintResult(t, testCases[tc].Pass, err2)
			} else {
				err3 := au.Auth(testCases[tc].id, methodCases[tc], &objectCases[tc])
				if err3 != nil {
					PrintResult(t, testCases[tc].Pass, err3)
				} else {
					PrintResult(t, testCases[tc].Pass, err2)
				}
			}
		}
	}
}

type TestInput1 struct {
	tip    TestInput
	id     string
	method object.Method
	key    object.Key
	Pass   bool
}

func TestHash(t *testing.T) {
	testCases := []TestInput1{
		{tip: TestInput{ConfigPath: cp[0]}, id: "agent-123", method: object.MethodGet, key: object.Key{Type: "world", Name: "judge", Namespace: "default"}, Pass: true},
		{tip: TestInput{ConfigPath: cp[1]}, id: "judge-123", method: object.MethodFind, key: object.Key{Type: "action"}, Pass: true},
		{tip: TestInput{ConfigPath: cp[2]}, id: "logger-123", method: object.MethodSet, key: object.Key{}, Pass: true},
		{tip: TestInput{ConfigPath: cp[6]}, id: "agent-123", method: object.MethodWatch, key: object.Key{Type: "action"}, Pass: false},
	}
	for tc, _ := range testCases {
		config, err1 := config.NewConfig(testCases[tc].tip.ConfigPath)
		if err1 != nil {
			PrintResult(t, testCases[tc].Pass, err1)
		} else {
			au, err2 := NewAuth(config)
			if err2 != nil {
				PrintResult(t, testCases[tc].Pass, err2)
			} else {
				hash0 := fmt.Sprintf("%s-%s-%s-%s-%s", testCases[tc].id, testCases[tc].method, testCases[tc].key.Type, testCases[tc].key.Namespace, testCases[tc].key.Name)
				hash1 := au.hash(testCases[tc].id, testCases[tc].method, testCases[tc].key)
				if hash0 != hash1 {
					fmt.Println(hash0, hash1, "sdf")
					PrintResult(t, testCases[tc].Pass, errors.New("The hash is not working properly"))
				} else {
					PrintResult(t, testCases[tc].Pass, nil)
				}
			}
		}
	}
}

var objectKeys = []object.Key{
	//config0
	object.Key{Type: "world", Name: "wrongName"},
	object.Key{Type: "action", Namespace: "default"},
	object.Key{Type: "action"},
	object.Key{Type: "world"},
	object.Key{},
	//config1
	object.Key{Type: "action", Namespace: "nondefault"},
	object.Key{Type: "world"},
	object.Key{Name: "wrongName"},
	//config2
	object.Key{Type: "nonaction", Name: "wrongName", Namespace: "nondefault"},
	object.Key{Type: "", Name: "wrongName", Namespace: "nondefault"},
	object.Key{Type: "action"},
	object.Key{Type: "nonaction", Name: "", Namespace: "nondefault"},
	//config3
	object.Key{Type: "nonaction", Namespace: "default"},
	object.Key{},
	object.Key{Type: "world"},

	object.Key{Type: "nonaction", Name: "wrongName", Namespace: "default"},
	object.Key{Type: "nonaction"},
	object.Key{Name: "wrongName", Namespace: "default"},
}
var objectCases = []object.Object{
	object.Object{Key: &objectKeys[0], Value: "Value"},
	object.Object{Key: &objectKeys[1], Value: "testing"},
	object.Object{Key: &objectKeys[2], Value: ""},
	object.Object{Key: &objectKeys[3], Value: "Value"},
	object.Object{Key: &objectKeys[4], Value: ""},
	object.Object{Key: &objectKeys[5], Value: "abc"},
	object.Object{Key: &objectKeys[6], Value: "abc"},
	object.Object{Key: &objectKeys[7], Value: ""},
	object.Object{Key: &objectKeys[8], Value: "Value"},
	object.Object{Key: &objectKeys[9], Value: ""},
	object.Object{Key: &objectKeys[10], Value: ""},
	object.Object{Key: &objectKeys[11], Value: "Value"},
	object.Object{Key: &objectKeys[12], Value: ""},
	object.Object{Key: &objectKeys[13], Value: ""},
	object.Object{Key: &objectKeys[14], Value: "Value"},
	object.Object{Key: &objectKeys[15], Value: "Value"},
	object.Object{Key: &objectKeys[16], Value: ""},
	object.Object{Key: &objectKeys[17], Value: ""},
}
var methodCases = []object.Method{
	object.MethodGet,
	object.MethodSet,
	object.MethodFind,
	object.MethodDelete,
	object.MethodWatch,
	object.MethodDelete,
	object.MethodSet,
	object.MethodFind,
	object.MethodDelete,
	object.MethodWatch,
	object.MethodGet,
	object.MethodSet,
	object.MethodSet,
	object.MethodFind,
	object.MethodWatch,
	object.MethodDelete,
	object.MethodFind,
}
