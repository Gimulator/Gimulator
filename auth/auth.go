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

type actor struct {
	username string
	password string
	role     string
}

type role struct {
	rules map[string]rule
}

type Auth struct {
	sync.Mutex
	path   string
	roles  map[string]role  // string is the name of Role
	actors map[string]actor // string is the username of actor
	log    *logrus.Entry
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
		Mutex:  sync.Mutex{},
		path:   path,
		roles:  make(map[string]role),
		actors: make(map[string]actor),
		log:    logrus.WithField("Entity", "auth"),
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

	actors, roles, err := loadConfig(a.path)
	if err != nil {
		a.log.WithError(err).Error("Can not read config file")
		return err
	}
	a.actors = actors
	a.roles = roles
	return nil
}

func (a *Auth) Authenticate(username, password string) (int, string) {
	a.log.Info("Start to Authenticate credential")

	actor, status, msg := a.getActor(username)
	if status != http.StatusOK {
		return status, msg
	}
	if actor.password != password {
		return http.StatusUnauthorized, fmt.Sprintf("Credential is not valid")
	}

	return http.StatusOK, ""
}

func (a *Auth) Authorize(name string, method Method, key *object.Key) (int, string) {
	a.log.Info("Start to authorize")
	actor, status, msg := a.getActor(name)
	if status != http.StatusOK {
		return http.StatusNotFound, msg
	}

	role, status, msg := a.getRole(actor.role)
	if status != http.StatusOK {
		return status, fmt.Sprintf("role '%s' does not exists", actor.role)
	}

	if rule, exists := role.rules[key.Type]; exists && rule.match(key, method) {
		return http.StatusOK, ""
	}
	return http.StatusUnauthorized, fmt.Sprintf("you don't have access on '%v', with method '%s' and role '%s'", *key, method, actor.role)
}

func (a *Auth) getActor(name string) (actor, int, string) {
	a.log.Info("Start to get actor")
	if actor, ex := a.actors[name]; ex {
		return actor, http.StatusOK, ""
	}
	return actor{}, http.StatusNotFound, fmt.Sprintf("Actor %s does not exist", name)
}

func (a *Auth) getRole(name string) (role, int, string) {
	a.log.Info("Start to get role")
	if r, ex := a.roles[name]; ex {
		return r, http.StatusOK, ""
	}
	return role{}, http.StatusNotFound, fmt.Sprintf("Role %s does not exist", name)
}
