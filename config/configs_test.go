package config

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

func TestLoadConfig(t *testing.T) {
	t.Logf("Given the need to test loadConfig method of Config type.")
	t.Logf("\tWhen checking the value \"%v\"", configPath)
	cWant := &Config{configPath, []Role{role1, role2, role3}, []Actor{actor1, actor2, actor3}, make(map[string]string), make(map[string][]Rule)}
	cGot := &Config{
		configPath:  configPath,
		idToRole:    make(map[string]string),
		roleToRules: make(map[string][]Rule),
	}
	err := cGot.loadConfig()

	if err != nil {
		t.Errorf(LogFailed(err, cWant, ballotX))
	} else if reflect.DeepEqual(cGot, cWant) {
		t.Logf(LogApproved(cWant, checkMark))
	} else {
		t.Errorf(LogFailed(cGot, cWant, ballotX))
	}
}

func TestValidateRoleRules(t *testing.T) {
	t.Logf("Given the need to test validateRolesRule method of Config type.")
	var tests = []struct {
		conf *Config
		want error
	}{
		{config1, nil},
		{config2, fmt.Errorf("error")},
	}

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v\"", test.conf)
		got := test.conf.validateRolesRule()

		if reflect.TypeOf(got) == reflect.TypeOf(test.want) {
			t.Logf(LogApproved(test.want, checkMark))
		} else {
			t.Errorf(LogFailed(got, test.want, ballotX))
		}
	}

}

func TestValidateActorsID(t *testing.T) {
	t.Logf("Given the need to test validateActorsID method of Config type.")
	var tests = []struct {
		conf *Config
		want error
	}{
		{config1, nil},
		{config3, fmt.Errorf("error")},
	}

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v\"", test.conf)
		got := test.conf.validateActorsID()

		if reflect.TypeOf(got) == reflect.TypeOf(test.want) {
			t.Logf(LogApproved(test.want, checkMark))
		} else {
			t.Errorf(LogFailed(got, test.want, ballotX))
		}
	}
}

func TestValidateActorsRole(t *testing.T) {
	t.Logf("Given the need to test validateActorsRole method of Config type.")
	var tests = []struct {
		conf *Config
		want error
	}{
		{config1, nil},
		{config4, fmt.Errorf("error")},
	}

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v\"", test.conf)
		got := test.conf.validateActorsRole()

		if reflect.TypeOf(got) == reflect.TypeOf(test.want) {
			t.Logf(LogApproved(test.want, checkMark))
		} else {
			t.Errorf(LogFailed(got, test.want, ballotX))
		}
	}
}

func TestGetRole(t *testing.T) {
	t.Logf("Given the need to test validateActorsRole method of Config type.")
	var tests = []struct {
		conf *Config
		id      string
		wantStr string
		wantErr error
	}{
		{config1, "agent-123", "agent", nil},
		{config1, "agent", "", fmt.Errorf("error")},
	}

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v %v\"", test.conf, test.id)
		gotStr, gotErr := test.conf.GetRole(test.id)

		if reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) && reflect.DeepEqual(gotStr, test.wantStr) {
			t.Logf(LogApproved(test.wantStr, checkMark))
		} else if !reflect.DeepEqual(gotStr, test.wantStr) {
			t.Errorf(LogFailed(gotStr, test.wantStr, ballotX))
		} else {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		}
	}
}

func TestGetRules(t *testing.T) {
	t.Logf("Given the need to test GetRules method of Config type.")
	var tests = []struct {
		conf *Config
		id      string
		wantRules []Rule
		wantErr error
	}{
		{config1, "agent-123", []Rule{rule1, rule2}, nil},
		{config1, "agent", nil, fmt.Errorf("error")},
	}

	for _, test := range tests {
		t.Logf("\tWhen checking the value \"%v %v\"", test.conf, test.id)
		gotRules, gotErr := test.conf.GetRules(test.id)

		if reflect.TypeOf(gotErr) == reflect.TypeOf(test.wantErr) && reflect.DeepEqual(gotRules, test.wantRules) {
			t.Logf(LogApproved(test.wantRules, checkMark))
		} else if !reflect.DeepEqual(gotRules, test.wantRules) {
			t.Errorf(LogFailed(gotRules, test.wantRules, ballotX))
		} else {
			t.Errorf(LogFailed(gotErr, test.wantErr, ballotX))
		}
	}
}

var (
	rule1 = Rule{Key : &object.Key{Type: "action", Namespace: "default"}, Methods : []object.Method{"set", "delete"}}
	rule2 = Rule{Key : &object.Key{Type: "world", Name: "judge", Namespace: "default"}, Methods : []object.Method{"get", "find", "watch"}}
	rule3 = Rule{Key : &object.Key{Type: "action"}, Methods : []object.Method{"get", "watch"}}
	rule4 = Rule{Key : &object.Key{Type: "world"}, Methods : []object.Method{"get", "set", "find", "watch", "delete"}}
	rule5 = Rule{Key : &object.Key{}, Methods : []object.Method{"get", "set", "find", "watch", "delete"}}

	role1 = Role{Name : "agent", Rules : []Rule{rule1, rule2}}
	role2 = Role{Name : "judge", Rules : []Rule{rule3, rule4}}
	role3 = Role{Name : "logger", Rules : []Rule{rule5}}

	actor1 = Actor{ID : "agent-123", Role : "agent"}
	actor2 = Actor{ID : "judge-123", Role : "judge"}
	actor3 = Actor{ID : "logger-123", Role : "logger"}

	idToRole1 = map[string]string{"agent-123" : "agent", "judge-123" : "judge", "logger-123" : "logger"}
	roleToRules1 = map[string][]Rule{"agent" : []Rule{rule1, rule2}, "judge" : []Rule{rule3, rule4}, "logger" : []Rule{rule5}}
	config1 = &Config{configPath, []Role{role1, role2, role3}, []Actor{actor1, actor2, actor3}, idToRole1, roleToRules1}

	idToRole2 = map[string]string{"agent-123" : "agent", "judge-123" : "judge", "logger-123" : "logger"}
	roleToRules2 = map[string][]Rule{"agent" : nil, "judge" : []Rule{rule3, rule4}, "logger" : []Rule{rule5}}
	config2 = &Config{configPathRoleWithoutRules, []Role{Role{Name:"agent"}, role2, role3}, []Actor{actor1, actor2, actor3}, idToRole1, roleToRules1}	

	idToRole3 = map[string]string{"agent-123" : "agent"}
	roleToRules3 = map[string][]Rule{"agent" : []Rule{rule1, rule2}, "judge" : []Rule{rule3, rule4}, "logger" : []Rule{rule5}}
	config3 = &Config{configPathDuplicateActors, []Role{Role{Name:"agent"}, role2, role3}, []Actor{actor1, actor1}, idToRole1, roleToRules1}	

	idToRole4 = map[string]string{"agent-123" : "agent", "judge-123" : "judge", "logger-123" : "logger"}
	roleToRules4 = map[string][]Rule{"judge" : []Rule{rule3, rule4}, "logger" : []Rule{rule5}}
	config4 = &Config{configPathInvalidRole, []Role{role2, role3}, []Actor{actor1, actor2, actor3}, idToRole1, roleToRules1}

	configPath                 = "configExamples/roles.yaml"
	configPathRoleWithoutRules = "configExamples/roles2.yaml"
	configPathDuplicateActors  = "configExamples/roles3.yaml"
	configPathInvalidRole      = "configExamples/roles4.yaml"
)