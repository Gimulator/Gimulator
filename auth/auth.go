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
	id           string
	role         string
	isRegistered bool
}

type role struct {
	rules map[string]*rule
}

type Auth struct {
	sync.Mutex
	path   string
	roles  map[string]*role  // string is the name of Role
	actors map[string]*actor // string is the id of actor
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
		roles:  make(map[string]*role),
		actors: make(map[string]*actor),
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

func (a *Auth) Register(id string) (int, string) {
	a.log.Info("Start to register new credential")

	actor, ex := a.actors[id]
	if !ex {
		return http.StatusNotFound, fmt.Sprintf("actor with id=%s does not exist", id)
	}
	actor.isRegistered = true

	return http.StatusOK, ""
}

func (a *Auth) Auth(id string, method Method, obj *object.Object) (int, string) {
	a.log.Info("Start to auth")

	actor, exists := a.actors[id]
	if !exists {
		return http.StatusNotFound, fmt.Sprintf("actor with id=%s does not exist", id)
	}

	if !actor.isRegistered {
		return http.StatusUnauthorized, fmt.Sprintf("you should first register")
	}

	if actor.id != obj.Owner {
		return http.StatusUnauthorized, fmt.Sprintf("id does not match with name=%s", obj.Owner)
	}

	role, exists := a.roles[actor.role]
	if !exists {
		return http.StatusNotFound, fmt.Sprintf("role '%s' does not exists", actor.role)
	}

	if rule, exists := role.rules[obj.Key.Type]; !exists || !rule.match(obj.Key, method) {
		return http.StatusUnauthorized, fmt.Sprintf("you don't have access on '%v', with method '%s' and role '%s'", *obj.Key, method, actor.role)
	}

	return http.StatusOK, ""
}
