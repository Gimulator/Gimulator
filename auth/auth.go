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

type UserManager struct {
	userStorage storage.UserStorage
	roleStorage storage.RoleStorage
}

func NewUserManager(credStorage storage.UserStorage, roleStorage storage.RoleStorage) (*UserManager, error) {
	return &UserManager{
		userStorage: credStorage,
		roleStorage: roleStorage,
	}, nil
}

func (a *UserManager) Authenticate(token string) (*storage.User, error) {
	user, err := a.userStorage.GetUserWithToken(token)
	if err == nil {
		return user, nil
	}

	if status.Code(err) == codes.NotFound {
		return user, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return user, err
}

func (a *UserManager) Authorize(role string, method types.Method, key *api.Key) error {
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

func (a *UserManager) UpdateStatus(id string, st types.Status) error {
	if err := a.userStorage.UpdateUserStatus(id, st); err != nil {
		return err
	}

	return nil
}

func (a *UserManager) GetStatus(id string) (types.Status, error) {
	user, err := a.userStorage.GetUserWithID(id)
	if err != nil {
		return types.StatusUnknown, err
	}

	return user.Status, nil
}

func (a *UserManager) validateDirectorAction(method types.Method, key *api.Key) error {
	keys, err := a.roleStorage.GetRules(string(types.DirectorRole), method)
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

func (a *UserManager) validateMasterAction(method types.Method, key *api.Key) error {
	return nil
}

func (a *UserManager) validateOperatorAction(method types.Method, key *api.Key) error {
	return nil
}

func (a *UserManager) validateActorAction(role string, method types.Method, key *api.Key) error {
	keys, err := a.roleStorage.GetRules(role, method)
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

func (a *UserManager) SetupMessage(token string, message *api.Message) error {
	user, err := a.userStorage.GetUserWithToken(token)
	if err != nil {
		return err
	}

	message.Meta = &api.Meta{
		CreationTime: timestamppb.Now(),
		Owner:        user.ID,
		Role:         user.Role,
	}

	return nil
}

func (a *UserManager) match(base, check *api.Key) bool {
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
