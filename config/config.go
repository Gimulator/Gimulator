package config

import (
	"os"
	"path/filepath"

	"github.com/Gimulator/protobuf/go/api"
	"gopkg.in/yaml.v2"
)

var (
	gimulatorConfigDir           string = "/etc/gimulator"
	gimulatorRulesFileName       string = "rules.yaml"
	gimulatorCredentialsFileName string = "credentials.yaml"
)

type Rule struct {
	Key     api.Key  `yaml:"key"`
	Methods []string `yaml:"methods"`
}

type Character struct {
	Director []Rule            `yaml:"director"`
	Actors   map[string][]Rule `yaml:"actors"`
	Master   []Rule            `yaml:"master,omitempty"`
	Operator []Rule            `yaml:"operator,omitempty"`
}

type Credential struct {
	Name      string `yaml:"name"`
	Token     string `yaml:"token"`
	Character string `yaml:"character"`
	Role      string `yaml:"role"`
}

type Config struct {
	Character   Character
	Credentials []Credential
}

func NewConfig(dir string) (*Config, error) {
	character, err := newCharacter(dir)
	if err != nil {
		return nil, err
	}

	creds, err := newCredentials(dir)
	if err != nil {
		return nil, err
	}

	return &Config{
		Character:   character,
		Credentials: creds,
	}, nil
}

func newCharacter(dir string) (Character, error) {
	if dir == "" {
		dir = gimulatorConfigDir
	}
	path := filepath.Join(dir, gimulatorRulesFileName)

	file, err := os.Open(path)
	if err != nil {
		return Character{}, err
	}

	character := Character{}
	if err := yaml.NewDecoder(file).Decode(&character); err != nil {
		return Character{}, err
	}

	character.Director = append(character.Director, Rule{
		Key: api.Key{},
		Methods: []string{
			api.Method_name[int32(api.Method_getActors)],
			api.Method_name[int32(api.Method_putResult)],
			api.Method_name[int32(api.Method_ping)],
		},
	})

	character.Operator = append(character.Operator, Rule{
		Key: api.Key{},
		Methods: []string{
			api.Method_name[int32(api.Method_setUserStatus)],
			api.Method_name[int32(api.Method_ping)],
		},
	})

	for i := range character.Actors {
		character.Actors[i] = append(character.Actors[i], Rule{
			Key: api.Key{},
			Methods: []string{
				api.Method_name[int32(api.Method_imReady)],
				api.Method_name[int32(api.Method_ping)],
			},
		})
	}

	return character, nil
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
