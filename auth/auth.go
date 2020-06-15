package auth

import (
	"fmt"
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
	log := logrus.WithField("entity", "auth")
	log.Info("starting to create new auth")

	fileInfo, err := os.Stat(path)
	if err != nil {
		log.WithError(err).Fatal("could not get stat of roles-file path")
		return nil, err
	}

	if fileInfo.IsDir() {
		log.Fatal("path is seted to a directory not file")
		return nil, fmt.Errorf("path must be a address of file not directory")
	}

	a := Auth{
		Mutex:  sync.Mutex{},
		path:   path,
		roles:  make(map[string]*role),
		actors: make(map[string]*actor),
		log:    log,
	}

	log.Info("starting to load configs")
	if err := a.loadConfigs(); err != nil {
		log.WithError(err).Error("could not load configs from roles file")
		return nil, err
	}

	return &a, nil
}

func (a *Auth) loadConfigs() error {
	a.Lock()
	a.Unlock()

	actors, roles, err := loadConfig(a.path)
	if err != nil {
		return err
	}
	a.actors = actors
	a.roles = roles
	return nil
}

func (a *Auth) Register(id string) error {
	actor, ex := a.actors[id]
	if !ex {
		return fmt.Errorf("actor with id=%s does not exist", id)
	}
	actor.isRegistered = true

	return nil
}

func (a *Auth) Auth(id string, method Method, obj *object.Object) error {
	actor, exists := a.actors[id]
	if !exists {
		return fmt.Errorf("actor with id=%s does not exist", id)
	}

	if !actor.isRegistered {
		return fmt.Errorf("actor with id=%s has not registerd", id)
	}

	if actor.id != id {
		return fmt.Errorf("id does not match with object owner")
	}

	role, exists := a.roles[actor.role]
	if !exists {
		// TODO: it's not a error from client: this should not be happend
		a.log.WithField("actor", actor).Fatal("role is seted by the auth package but there is no role with specified actor.role")
		return fmt.Errorf("role '%s' does not exists", actor.role)
	}

	if rule, exists := role.rules[obj.Key.Type]; exists && rule.match(obj.Key, method) {
		return nil
	}

	if rule, exists := role.rules[""]; exists && rule.match(obj.Key, method) {
		return nil
	}

	// TODO: is this Ok to return actor.role to the client???
	return fmt.Errorf("access denied on key=%s, method=%s, role=%s", *obj.Key, method, actor.role)
}
