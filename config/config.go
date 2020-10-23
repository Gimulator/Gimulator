package config

import (
	"os"

	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"gopkg.in/yaml.v2"
)

var (
	gimulatorRolesDefaultPath       string = "/etc/gimulator/roles.yaml"
	gimulatorCredentialsDefaultPath string = "/etc/gimulator/credentials.yaml"
)

type Rule struct {
	Key     *api.Key       `yaml:"key"`
	Methods []types.Method `yaml:"methods"`
}

type Roles struct {
	Director []Rule            `yaml:"director"`
	Actors   map[string][]Rule `yaml:"actors"`
}

type Credential struct {
	Token string `yaml:"token"`
	ID    string `yaml:"id"`
	Role  string `yaml:"role"`
}

func NewRoles(path string) (*Roles, error) {
	if path == "" {
		path = gimulatorRolesDefaultPath
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	roles := &Roles{}
	if err := yaml.NewDecoder(file).Decode(roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func NewCredentials(path string) ([]Credential, error) {
	if path == "" {
		path = gimulatorCredentialsDefaultPath
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	credentials := []Credential{}
	if err := yaml.NewDecoder(file).Decode(&credentials); err != nil {
		return nil, err
	}

	return credentials, nil
}
