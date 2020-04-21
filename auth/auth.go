package auth

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/Gimulator/Gimulator/object"
	"github.com/sirupsen/logrus"
)

type Method string

const (
	Get    Method = "get"
	Set    Method = "set"
	Find   Method = "find"
	Delete Method = "delete"
	Watch  Method = "watch"
)

type rule struct {
	key     object.Key
	methods map[Method]bool
}

func (r *rule) match(key *object.Key, method Method) bool {
	if !r.key.Match(key) {
		return false
	}

	if _, exists := r.methods[method]; exists {
		return true
	}
	return false
}

type role struct {
	role     string
	password string          //`yaml:"password"`
	rules    map[string]rule //`yaml:"types"`
}

type Auth struct {
	sync.Mutex
	path  string
	roles map[string]role //`yaml:"roles"`
	log   *logrus.Entry
}

func NewAuth(path string) (*Auth, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path must be a address of file not directory")
	}

	a := Auth{
		Mutex: sync.Mutex{},
		path:  path,
		roles: make(map[string]role),
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

	if err := readConfig(a.path, &a.roles); err != nil {
		a.log.WithError(err).Error("Can not read config file")
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
	if role.password != password {
		return http.StatusUnauthorized, fmt.Sprintf("Credential is not valid")
	}

	return http.StatusOK, ""
}

func (a *Auth) Authorize(role string, method Method, key *object.Key) (int, string) {
	a.log.Info("Start to authorize")
	if actualRole, exists := a.roles[role]; exists {
		if actualType, exists := actualRole.rules[key.Type]; exists && actualType.match(key, method) {
			return http.StatusOK, ""
		}
		return http.StatusUnauthorized, fmt.Sprintf("you don't have access on '%v', with method '%s' and role '%s'", *key, method, role)
	}
	return http.StatusNotFound, fmt.Sprintf("role '%s' does not exists", role)
}

func (a *Auth) getRole(r string) (role, int, string) {
	a.log.Info("Start to get role")
	if r, ex := a.roles[r]; ex {
		return r, http.StatusOK, ""
	}
	return role{}, http.StatusNotFound, fmt.Sprintf("Role %s does not exist", r)
}
