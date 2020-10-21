package auth

import (
	"time"

	"github.com/Gimulator/Gimulator/types.go"
	"github.com/Gimulator/protobuf/go/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultExpirationTime  = time.Second * 120
	defaultCleanupInterval = time.Second * 180
)

type Storage interface {
	GetRole(id string) string
	GetRules(role string, method types.Method, key *api.Key) []string
}

type Auther struct {
	storage Storage
}

func NewAuther(storage Storage) (*Auther, error) {
	return &Auther{
		storage: storage,
	}, nil
}

func (a *Auther) Auth(id string, method types.Method, key *api.Key) error {
	role := a.storage.GetRole(id)
	if role == "" {
		return status.Errorf(codes.Unauthenticated, "couldn't role based on id")
	}

	switch role {
	case string(types.DirectorRole):
		return a.validateDirectorAction(id, role, method, key)
	case string(types.MasterRole):
		return a.validateMasterAction(id, role, method, key)
	case string(types.OperatorRole):
		return a.validateOperatorAction(id, role, method, key)
	default:
		return a.validateActorAction(id, role, method, key)
	}
}

func (a *Auther) validateDirectorAction(id, role string, method types.Method, key *api.Key) error {
	rules := a.storage.GetRules(role, method, key)
	if len(rules) == 0 {
		return status.Errorf(codes.PermissionDenied, "")
	}
	return nil
}

func (a *Auther) validateMasterAction(id, role string, method types.Method, key *api.Key) error {
	return nil
}

func (a *Auther) validateOperatorAction(id, role string, method types.Method, key *api.Key) error {
	return nil
}

func (a *Auther) validateActorAction(id, role string, method types.Method, key *api.Key) error {
	rules := a.storage.GetRules(role, method, key)
	if len(rules) == 0 {
		return status.Errorf(codes.PermissionDenied, "")
	}
	return nil
}
