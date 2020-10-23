package auth

import (
	"time"

	"github.com/Gimulator/Gimulator/storage"
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

type Auther struct {
	storage storage.AuthStorage
}

func NewAuther(storage storage.AuthStorage) (*Auther, error) {
	return &Auther{
		storage: storage,
	}, nil
}

func (a *Auther) Authenticate(token string) (string, string, error) {
	id, role, err := a.storage.GetCredWithToken(token)
	if err == nil {
		return id, role, nil
	}

	if status.Code(err) == codes.NotFound {
		return "", "", status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return "", "", err
}

func (a *Auther) Authorize(role string, method types.Method, key *api.Key) error {
	if role == "" {
		return status.Errorf(codes.Unauthenticated, "couldn't find role based on id")
	}

	switch role {
	case string(types.DirectorRole):
		return a.validateDirectorAction(method, key)
	case string(types.MasterRole):
		return a.validateMasterAction(method, key)
	case string(types.OperatorRole):
		return a.validateOperatorAction(method, key)
	default:
		return a.validateActorAction(role, method, key)
	}
}

func (a *Auther) validateDirectorAction(method types.Method, key *api.Key) error {
	keys, err := a.storage.GetRules(string(types.DirectorRole), method)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if a.match(base, key) {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "")
}

func (a *Auther) validateMasterAction(method types.Method, key *api.Key) error {
	return nil
}

func (a *Auther) validateOperatorAction(method types.Method, key *api.Key) error {
	return nil
}

func (a *Auther) validateActorAction(role string, method types.Method, key *api.Key) error {
	keys, err := a.storage.GetRules(role, method)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if a.match(base, key) {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "")
}

func (a *Auther) SetupMessage(token string, message *api.Message) error {
	role, id, err := a.storage.GetCredWithToken(token)
	if err != nil {
		return err
	}

	message.Meta = &api.Meta{
		CreationTime: timestamppb.Now(),
		Owner:        id,
		Role:         role,
	}

	return nil
}

func (a *Auther) match(base, check *api.Key) bool {
	if base.Type != "" && base.Type != check.Type {
		return false
	}
	if base.Name != "" && base.Name != check.Name {
		return false
	}
	if base.Namespace != "" && base.Namespace != check.Namespace {
		return false
	}
	return true
}
