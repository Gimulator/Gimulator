package auth

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"gitlab.com/Syfract/Xerac/gimulator/object"
	"gopkg.in/yaml.v2"
)

type Method string

const (
	Get    Method = "get"
	Set    Method = "set"
	Find   Method = "find"
	Delete Method = "delete"
	Watch  Method = "watch"
)

type Type struct {
	Key     object.Key `json:"key"`
	Methods []Method   `json:"methods"`
}

func (t *Type) match(key object.Key, method Method) bool {
	if !t.Key.Match(key) {
		return false
	}
	for _, m := range t.Methods {
		if m == method {
			return true
		}
	}
	return false
}

type Role struct {
	Password string          `json:"password"`
	Types    map[string]Type `json:"types"`
}

type Auth struct {
	sync.Mutex
	path          string
	Roles         map[string]Role `json:"roles"`
	tokenToClient map[string]*Client
	nameToClient  map[string]*Client
}

func NewAuth(path string) (*Auth, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path must be a path to file not directory")
	}

	a := Auth{
		Mutex:         sync.Mutex{},
		path:          path,
		Roles:         nil,
		nameToClient:  make(map[string]*Client),
		tokenToClient: make(map[string]*Client),
	}

	if err := a.loadConfigs(); err != nil {
		return nil, err
	}

	return &a, nil
}

func (a *Auth) loadConfigs() error {
	a.Lock()
	a.Unlock()

	file, err := os.Open(a.path)
	if err != nil {
		return err
	}

	if err := yaml.NewDecoder(file).Decode(&a.Roles); err != nil {
		return err
	}
	return nil
}

func (a *Auth) Authenticate(cred Credential) (int, string) {
	role, status, msg := a.getRole(cred.Role)
	if status != http.StatusAccepted {
		return status, msg
	}
	if role.Password != cred.Password {
		return http.StatusUnauthorized, fmt.Sprintf("Credential is not valid")
	}

	return http.StatusAccepted, ""
}

func (a *Auth) GetClientWithName(name string) (*Client, int, string) {
	if cli, exist := a.nameToClient[name]; exist {
		return cli, http.StatusAccepted, ""
	}
	return nil, http.StatusNotFound, fmt.Sprintf("User with name '%s' does not exist", name)
}

func (a *Auth) GetClientWithToken(token string) (*Client, int, string) {
	if cli, exist := a.tokenToClient[token]; exist {
		return cli, http.StatusAccepted, ""
	}
	return nil, http.StatusNotFound, fmt.Sprintf("User with token '%s' does not exist", token)
}

func (a *Auth) CreateNewClient(cred Credential) (*Client, int, string) {
	a.Lock()
	defer a.Unlock()

	token, status := newCookie()
	if status != http.StatusAccepted {
		return nil, status, "Can not generate new cookie"
	}

	client := NewClient(cred, token)
	a.tokenToClient[token] = client
	a.nameToClient[cred.Username] = client

	return client, http.StatusAccepted, ""
}

func (a *Auth) Authorize(cli *Client, key object.Key, method Method) (int, string) {
	role := cli.cred.Role
	if actualRole, exists := a.Roles[role]; exists {
		if actualType, exists := actualRole.Types[key.Type]; exists && actualType.match(key, method) {
			return http.StatusAccepted, ""
		}
		return http.StatusUnauthorized, fmt.Sprintf("you don't have access on '%v', with method '%s' and role %s", key, method, role)
	}
	return http.StatusNotFound, fmt.Sprintf("role '%s' does not exists", role)
}

func (a *Auth) HandleRequest(w http.ResponseWriter, r *http.Request, method Method, cli *Client, obj *object.Object) (int, string) {
	var (
		token  string
		status int
		msg    string
		client *Client
		object object.Object
	)

	token, status, msg = getCookie(r)
	if status != http.StatusAccepted {
		return status, msg
	}

	client, status, msg = a.GetClientWithToken(token)
	if status != http.StatusAccepted {
		return status, msg
	}

	if status, msg := decodeJSONBody(w, r, &object); status != http.StatusAccepted {
		return status, msg
	}

	if status, msg := a.Authorize(client, object.Key, method); status != http.StatusAccepted {
		return status, msg
	}

	*cli = *client
	*obj = object
	return http.StatusAccepted, ""
}

func (a *Auth) RegisterNewClient(w http.ResponseWriter, r *http.Request, cli *Client, obj *object.Object) (int, string) {
	var cred Credential
	status, msg := decodeJSONBody(w, r, &cred)
	if status != http.StatusAccepted {
		return status, msg
	}
	status, msg = a.Authenticate(cred)
	if status != http.StatusAccepted {
		return status, msg
	}

	var client *Client
	client, status, msg = a.GetClientWithName(cred.Username)
	if status == http.StatusAccepted {
		*cli = *client
		return http.StatusAccepted, ""
	}

	client, status, msg = a.CreateNewClient(cred)
	if status != http.StatusAccepted {
		return status, msg
	}

	*cli = *client
	return http.StatusAccepted, ""
}

func (a *Auth) getRole(role string) (Role, int, string) {
	if r, ex := a.Roles[role]; ex {
		return r, http.StatusAccepted, ""
	}
	return Role{}, http.StatusNotFound, fmt.Sprintf("Role %s does not exist", role)
}
