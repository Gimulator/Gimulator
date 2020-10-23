package auth

import (
	"time"

	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultExpirationTime  = time.Second * 120
	defaultCleanupInterval = time.Second * 180
)

type Storage interface {
	GetRoleWithToken(string) string
	GetIDWithToken(string) string
	GetRoleIDWithToken(string) (string, string)
	GetRules(string, types.Method, *api.Key) []string
}

type Auther struct {
	storage Storage
}

func NewAuther(storage Storage) (*Auther, error) {
	return &Auther{
		storage: storage,
	}, nil
}

func (a *Auther) Auth(token string, method types.Method, key *api.Key) error {
	role := a.storage.GetRoleWithToken(token)
	if role == "" {
		return status.Errorf(codes.Unauthenticated, "couldn't find role based on id")
	}

	switch role {
	case string(types.DirectorRole):
		return a.validateDirectorAction(token, role, method, key)
	case string(types.MasterRole):
		return a.validateMasterAction(token, role, method, key)
	case string(types.OperatorRole):
		return a.validateOperatorAction(token, role, method, key)
	default:
		return a.validateActorAction(token, role, method, key)
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

func (a *Auther) SetupMessage(token string, message *api.Message) {
	role, id := a.storage.GetRoleIDWithToken(token)
	meta := &api.Meta{
		CreationTime: timestamppb.Now(),
		Owner:        id,
		Role:         role,
	}

	message.Meta = meta
}
