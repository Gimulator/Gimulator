package config

import (
	"os"
	"path/filepath"

	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"gopkg.in/yaml.v2"
)

var (
	gimulatorConfigDir           string = "/etc/gimulator"
	gimulatorRolesFileName       string = "roles.yaml"
	gimulatorCredentialsFileName string = "credentials.yaml"
)

type Rule struct {
	Key     api.Key        `yaml:"key"`
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

type Config struct {
	Roles       Roles
	Credentials []Credential
}

func NewConfig(dir string) (*Config, error) {
	roles, err := newRoles(dir)
	if err != nil {
		return nil, err
	}

	creds, err := newCredentials(dir)
	if err != nil {
		return nil, err
	}

	return &Config{
		Roles:       roles,
		Credentials: creds,
	}, nil
}

func newRoles(dir string) (Roles, error) {
	if dir == "" {
		dir = gimulatorConfigDir
	}
	path := filepath.Join(dir, gimulatorRolesFileName)

	file, err := os.Open(path)
	if err != nil {
		return Roles{}, err
	}

	roles := Roles{}
	if err := yaml.NewDecoder(file).Decode(&roles); err != nil {
		return Roles{}, err
	}

	return roles, nil
}

func newCredentials(dir string) ([]Credential, error) {
	if dir == "" {
		dir = gimulatorConfigDir
	}
	path := filepath.Join(dir, gimulatorCredentialsFileName)

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
