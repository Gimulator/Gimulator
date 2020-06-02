package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/Gimulator/Gimulator/object"
	"gopkg.in/yaml.v2"
)

type Rule struct {
	Type      string   `yaml:"type"`
	Name      string   `yaml:"name"`
	Namespace string   `yaml:"namespace"`
	Methods   []Method `yaml:"methods"`
}

type Role struct {
	Role  string `yaml:"role"`
	Rules []Rule `yaml:"rules"`
}

type Actor struct {
	ID   string `yaml:"id"`
	Role string `yaml:"role"`
}

type Config struct {
	Roles  []Role  `yaml:"roles"`
	Actors []Actor `yaml:"actors"`
}

func loadConfig(path string) (map[string]*actor, map[string]*role, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, nil, err
	}

	roles := loadRoles(config.Roles)
	actors := loadActors(config.Actors)
	err = check_actors(actors, roles)
	return actors, roles, err
}

func check_actors(actors map[string]*actor, roles map[string]*role) error {
	for _, val := range actors {
		flag := true
		for key := range roles {
			if val.role == key {
				flag = false
			}
		}
		if flag {
			return errors.New(fmt.Sprintf("Actor with role %s is invalid.", val.role))
		}
	}
	return nil
}

func loadRoles(cRoles []Role) map[string]*role {
	aRoles := make(map[string]*role)
	for _, cRole := range cRoles {
		aRoles[cRole.Role] = loadRole(cRole)
	}
	return aRoles
}

func loadRole(cRole Role) *role {
	return &role{
		rules: loadRules(cRole.Rules),
	}
}

func loadRules(cRules []Rule) map[string]*rule {
	aRules := make(map[string]*rule)
	for _, cRule := range cRules {
		aRules[cRule.Type] = loadRule(cRule)
	}

	return aRules
}

func loadRule(src Rule) *rule {
	dst := &rule{}
	dst.key = object.Key{
		Name:      src.Name,
		Namespace: src.Namespace,
		Type:      src.Type,
	}

	dst.methods = make(map[Method]bool)
	for _, m := range src.Methods {
		dst.methods[m] = true
	}

	return dst
}

func loadActors(cActors []Actor) map[string]*actor {
	aActor := make(map[string]*actor)
	for _, cActor := range cActors {
		aActor[cActor.ID] = loadActor(cActor)
	}
	return aActor

}

func loadActor(cActor Actor) *actor {
	return &actor{
		id:           cActor.ID,
		role:         cActor.Role,
		isRegistered: false,
	}
}
