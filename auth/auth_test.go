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
	for _, test := range testCases {
		config, err := config.NewConfig(test.ConfigPath)
		if err != nil {
			PrintResult(t, test.Pass, err)
			continue
		}
		if _, err := NewAuth(config); err != nil {
			PrintResult(t, test.Pass, err)
			continue
		}
		PrintResult(t, test.Pass, nil)
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

	for _, test := range testCases {
		config, err := config.NewConfig(test.ConfigPath)
		if err != nil {
			PrintResult(t, test.Pass, err)
			continue
		}
		au, err := NewAuth(config)
		if err != nil {
			PrintResult(t, test.Pass, err)
			continue
		}

		if err := au.Register(test.id); err != nil {
			PrintResult(t, test.Pass, err)
			continue
		}
		PrintResult(t, test.Pass, nil)
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
	for tc := range testCases {
		config, err := config.NewConfig(testCases[tc].ConfigPath)
		if err != nil {
			PrintResult(t, testCases[tc].Pass, err)
			continue
		}
		au, err := NewAuth(config)
		if err != nil {
			PrintResult(t, testCases[tc].Pass, err)
			continue
		}

		if err := au.Auth(testCases[tc].id, methodCases[tc], &objectCases[tc]); err != nil {
			PrintResult(t, testCases[tc].Pass, err)
			continue
		}
		PrintResult(t, testCases[tc].Pass, nil)
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
		{tip: TestInput{ConfigPath: cp[0]}, id: "agent-123", method: methodCases[0], key: object.Key{Type: "world", Name: "judge", Namespace: "default"}, Pass: true},
		{tip: TestInput{ConfigPath: cp[1]}, id: "judge-123", method: methodCases[2], key: object.Key{Type: "action"}, Pass: true},
		{tip: TestInput{ConfigPath: cp[2]}, id: "logger-123", method: methodCases[1], key: object.Key{}, Pass: true},
		{tip: TestInput{ConfigPath: cp[6]}, id: "agent-123", method: methodCases[4], key: object.Key{Type: "action"}, Pass: false},
	}
	for _, test := range testCases {
		config, err := config.NewConfig(test.tip.ConfigPath)
		if err != nil {
			PrintResult(t, test.Pass, err)
			continue
		}
		au, err := NewAuth(config)
		if err != nil {
			PrintResult(t, test.Pass, err)
			continue
		}
		hash0 := fmt.Sprintf("%s-%s-%s-%s-%s", test.id, test.method, test.key.Type, test.key.Namespace, test.key.Name)
		hash1 := au.hash(test.id, test.method, test.key)
		if hash0 != hash1 {
			PrintResult(t, test.Pass, errors.New("The hash is not working properly"))
			continue
		}
		PrintResult(t, test.Pass, nil)
	}
}

var objectKeys = []object.Key{
	//config0
	{Type: "world", Name: "wrongName"},
	{Type: "action", Namespace: "default"},
	{Type: "action"},
	{Type: "world"},
	{},
	//config1
	{Type: "action", Namespace: "nondefault"},
	{Type: "world"},
	{Name: "wrongName"},
	//config2
	{Type: "nonaction", Name: "wrongName", Namespace: "nondefault"},
	{Type: "", Name: "wrongName", Namespace: "nondefault"},
	{Type: "action"},
	{Type: "nonaction", Name: "", Namespace: "nondefault"},
	//config3
	{Type: "nonaction", Namespace: "default"},
	{},
	{Type: "world"},
	//others
	{Type: "nonaction", Name: "wrongName", Namespace: "default"},
	{Type: "nonaction"},
	{Name: "wrongName", Namespace: "default"},
}
var objectCases = []object.Object{
	//object.Object{},    // expected to fail but the program exits
	{Key: &objectKeys[0], Value: "Value"},
	{Key: &objectKeys[1], Value: "testing"},
	{Key: &objectKeys[2], Value: ""},
	{Key: &objectKeys[3], Value: "Value"},
	{Key: &objectKeys[4], Value: ""},
	{Key: &objectKeys[5], Value: "abc"},
	{Key: &objectKeys[6], Value: "abc"},
	{Key: &objectKeys[7], Value: ""},
	{Key: &objectKeys[8], Value: "Value"},
	{Key: &objectKeys[9], Value: ""},
	{Key: &objectKeys[10], Value: ""},
	{Key: &objectKeys[11], Value: "Value"},
	{Key: &objectKeys[12], Value: ""},
	{Key: &objectKeys[13], Value: ""},
	{Key: &objectKeys[14], Value: "Value"},
	{Key: &objectKeys[15], Value: "Value"},
	{Key: &objectKeys[16], Value: ""},
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
