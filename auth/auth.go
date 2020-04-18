package auth

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/Gimulator/Gimulator/object"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Method string

const (
	Get    Method = "get"
	Set    Method = "set"
	Find   Method = "find"
	Delete Method = "delete"
	Watch  Method = "watch"
)

type Type struct {
	Key     object.Key
	Methods []Method
}

func (t *Type) match(key *object.Key, method Method) bool {
	if !t.Key.Match(key) {
		return false
	}
	for _, m := range t.Methods {
		if m == method {
			return true
		}
	}
	return false
}

type Role struct {
	Password string          //`yaml:"password"`
	Types    map[string]Type //`yaml:"types"`
}

type Auth struct {
	sync.Mutex
	path  string
	Roles map[string]Role //`yaml:"roles"`
	log   *logrus.Entry
}

func NewAuth(path string) (*Auth, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path must be a path to file not directory")
	}

	a := Auth{
		Mutex: sync.Mutex{},
		path:  path,
		Roles: make(map[string]Role),
		log:   logrus.WithField("Entity", "auth"),
	}

	if err := a.loadConfigs(); err != nil {
		return nil, err
	}

	return &a, nil
}

func (a *Auth) loadConfigs() error {
	a.log.Info("Start to load config")
	a.Lock()
	a.Unlock()

	file, err := os.Open(a.path)
	if err != nil {
		a.log.WithError(err).Error("Can not open config file")
		return err
	}

	if err := yaml.NewDecoder(file).Decode(&a); err != nil {
		a.log.WithError(err).Error("Can not decode config file")
		return err
	}

	return nil
}

func (a *Auth) Authenticate(roleName, password string) (int, string) {
	a.log.Info("Start to Authenticate credential")

	role, status, msg := a.getRole(roleName)
	if status != http.StatusOK {
		return status, msg
	}
	if role.Password != password {
		return http.StatusUnauthorized, fmt.Sprintf("Credential is not valid")
	}

	return http.StatusOK, ""
}

func (a *Auth) Authorize(role string, method Method, key *object.Key) (int, string) {
	a.log.Info("Start to authorize")
	if actualRole, exists := a.Roles[role]; exists {
		if actualType, exists := actualRole.Types[key.Type]; exists && actualType.match(key, method) {
			return http.StatusOK, ""
		}
		return http.StatusUnauthorized, fmt.Sprintf("you don't have access on '%v', with method '%s' and role %s", key, method, role)
	}
	return http.StatusNotFound, fmt.Sprintf("role '%s' does not exists", role)
}

func (a *Auth) getRole(role string) (Role, int, string) {
	a.log.Info("Start to get role")
	if r, ex := a.Roles[role]; ex {
		return r, http.StatusOK, ""
	}
	return Role{}, http.StatusNotFound, fmt.Sprintf("Role %s does not exist", role)
}
