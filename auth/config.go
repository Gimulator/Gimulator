package auth

import (
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
	Role     string `yaml:"role"`
	Password string `yaml:"password"`
	Rules    []Rule `yaml:"rules"`
}

type Config struct {
	Roles []Role `yaml:"roles"`
}

func readConfig(path string, dst *map[string]role) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return err
	}

	*dst = loadRoles(config.Roles)
	return nil
}

func loadRoles(src []Role) map[string]role {
	dst := make(map[string]role)
	for _, Role := range src {
		dst[Role.Role] = loadRole(Role)
	}
	return dst
}

func loadRole(src Role) role {
	dst := role{}
	dst.password = src.Password

	dst.rules = make(map[string]rule)
	for _, rule := range src.Rules {
		dst.rules[rule.Type] = loadRule(rule)
	}
	return dst
}

func loadRule(src Rule) rule {
	dst := rule{}
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
