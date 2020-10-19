package config

import (
	"os"

	"github.com/Gimulator/protobuf/go/api"
	"gopkg.in/yaml.v2"
)

var (
	gimulatorRolesDefaultPath       string = "/etc/gimulator/roles.yaml"
	gimulatorCredentialsDefaultPath string = "/etc/gimulator/credentials.yaml"
)

type Rule struct {
	Key     *api.Key `json:"key"`
	Methods []string `json:"methods"`
}

type Roles struct {
	Director []Rule            `json:"director"`
	Actors   map[string][]Rule `json:"actors"`
}

type Credentials struct {
	Roles map[string]string `json:"roles"`
}

func newRoles(path string) (*Roles, error) {
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

func NewCredentials(path string) (*Credentials, error) {
	if path == "" {
		path = gimulatorCredentialsDefaultPath
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	credentials := &Credentials{}
	if err := yaml.NewDecoder(file).Decode(credentials); err != nil {
		return nil, err
	}

	return credentials, nil
}
