package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/object"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

const (
	defaultExpirationTime  = time.Second * 120
	defaultCleanupInterval = time.Second * 180
)

type response struct {
	isAllowed bool
	err       error
}

type Auth struct {
	sync.Mutex

	config *config.Config
	cache  *cache.Cache
	log    *logrus.Entry
}

func NewAuth(config *config.Config) (*Auth, error) {
	log := logrus.WithField("entity", "auth")
	log.Info("starting to create new auth")

	return &Auth{
		Mutex:  sync.Mutex{},
		config: config,
		cache:  cache.New(defaultExpirationTime, defaultCleanupInterval),
		log:    log,
	}, nil
}

func (a *Auth) Register(id string) error {
	if err := a.config.DoesIdExist(id); err != nil {
		return err
	}
	return nil
}

func (a *Auth) Auth(id string, method object.Method, obj *object.Object) error {
	if obj.Key == nil {
		return fmt.Errorf("object's key is empty")
	}
	hash := a.hash(id, method, *obj.Key)

	if i, exists := a.cache.Get(hash); exists {
		resp := i.(response)
		if resp.isAllowed {
			return nil
		}
		return resp.err
	}

	resp := response{
		isAllowed: false,
		err:       nil,
	}
	defer func() {
		a.cache.Add(hash, resp, defaultExpirationTime)
	}()

	err := a.config.DoesIdExist(id)
	if err != nil {
		resp.err = err
		return err
	}

	rules, err := a.config.GetRules(id)
	if err != nil {
		resp.err = err
		return err
	}

	for _, rule := range rules {
		if !rule.Key.Match(obj.Key) {
			continue
		}

		for _, met := range rule.Methods {
			if met != method {
				continue
			}
			resp.isAllowed = true
			resp.err = nil
			return nil
		}
	}

	err = fmt.Errorf("access denied on id='%s' key='%v', method='%s'", id, *obj.Key, method)
	resp.err = err
	resp.isAllowed = false

	return err
}

func (a *Auth) getResp(hash string) (response, bool) {
	i, exists := a.cache.Get(hash)
	return i.(response), exists
}

func (a *Auth) hash(id string, method object.Method, key object.Key) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", id, method, key.Type, key.Namespace, key.Name)
}
