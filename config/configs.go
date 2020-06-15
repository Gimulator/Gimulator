package config

import (
	"fmt"
	"os"

	"github.com/Gimulator/Gimulator/object"
	"gopkg.in/yaml.v2"
)

type Rule struct {
	Key     object.Key      `yaml:"key"`
	Methods []object.Method `yaml:"methods"`
}

type Role struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

type Actor struct {
	ID   string `yaml:"id"`
	Role string `yaml:"role"`
}

type Config struct {
	configPath  string
	Roles       []Role  `yaml:"roles"`
	Actors      []Actor `yaml:"actors"`
	idToRole    map[string]string
	roleToRules map[string][]Rule
}

func NewConfig(path string) (*Config, error) {
	c := &Config{
		configPath:  path,
		idToRole:    make(map[string]string),
		roleToRules: make(map[string][]Rule),
	}

	if err := c.loadConfig(); err != nil {
		return nil, err
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) loadConfig() error {
	file, err := os.Open(c.configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return err
	}

	return nil
}

func (c *Config) validate() error {
	if err := c.validateRoles(); err != nil {
		return err
	}

	if err := c.validateActors(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateRoles() error {
	if err := c.validateRolesName(); err != nil {
		return err
	}

	if err := c.validateRolesRule(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateRolesName() error {
	tmp := make(map[string]bool)
	for _, role := range c.Roles {
		tmp[role.Name] = true
	}

	if len(tmp) != len(c.Roles) {
		return fmt.Errorf("duplicate role is not allowed")
	}

	return nil
}

func (c *Config) validateRolesRule() error {
	for _, role := range c.Roles {
		rules := role.Rules
		if rules == nil || len(rules) == 0 {
			return fmt.Errorf("empty list of rules, for role '%s', is not allowed", role.Name)
		}
	}
	return nil
}

func (c *Config) validateActors() error {
	if err := c.validateActorsRole(); err != nil {
		return err
	}

	if err := c.validateActorsID(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateActorsID() error {
	tmp := make(map[string]bool)

	for _, actor := range c.Actors {
		tmp[actor.ID] = true
	}

	if len(tmp) != len(c.Actors) {
		return fmt.Errorf("duplicate actor is not allowed")
	}

	return nil
}

func (c *Config) validateActorsRole() error {
	for _, actor := range c.Actors {
		actorRole := actor.Role

		isValid := false
		for _, role := range c.Roles {
			if actorRole == role.Name {
				isValid = true
			}
		}

		if !isValid {
			return fmt.Errorf("actor '%s' has invalid role '%s'", actor.ID, actor.Role)
		}
	}
	return nil
}

func (c *Config) postprocess() {
	for _, actor := range c.Actors {
		c.idToRole[actor.ID] = actor.Role
	}

	for _, role := range c.Roles {
		c.roleToRules[role.Name] = role.Rules
	}
}

func (c *Config) GetRole(id string) (string, error) {
	if role, exists := c.idToRole[id]; exists {
		return role, nil
	}
	return "", fmt.Errorf("id '%s' is not found", id)
}

func (c *Config) GetRules(id string) ([]Rule, error) {
	role, err := c.GetRole(id)
	if err != nil {
		return nil, err
	}

	return c.roleToRules[role], nil
}

func (c *Config) DoesIdExist(id string) bool {
	if _, exists := c.idToRole[id]; exists {
		return true
	}
	return false
}
