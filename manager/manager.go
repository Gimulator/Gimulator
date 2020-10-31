package manager

import (
	"github.com/Gimulator/Gimulator/storage"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Manager struct {
	userStorage storage.UserStorage
	roleStorage storage.RoleStorage
}

func NewManager(credStorage storage.UserStorage, roleStorage storage.RoleStorage) (*Manager, error) {
	return &Manager{
		userStorage: credStorage,
		roleStorage: roleStorage,
	}, nil
}

func (m *Manager) Authenticate(token string) (*api.User, error) {
	user, err := m.userStorage.GetUserWithToken(token)
	if err == nil {
		return user, nil
	}

	if status.Code(err) == codes.NotFound {
		return user, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return user, err
}

func (m *Manager) UpdateStatus(id string, st api.Status) error {
	return m.userStorage.UpdateUserStatus(id, st)
}

func (m *Manager) UpdateReadiness(id string, isReady bool) error {
	return m.userStorage.UpdateUserReadiness(id, isReady)
}

func (m *Manager) GetUserWithID(id string) (*api.User, error) {
	return m.userStorage.GetUserWithID(id)
}

func (m *Manager) GetUsersWithCharacter(character api.Character) ([]*api.User, error) {
	return m.userStorage.GetUsersWithCharacter(character)
}

func (m *Manager) Authorize(user *api.User, method api.Method, info interface{}) error {
	switch user.Character {
	case api.Character_director:
		return m.validateDirectorAction(user, method, info)
	case api.Character_master:
		return m.validateMasterAction(user, method, info)
	case api.Character_operator:
		return m.validateOperatorAction(user, method, info)
	case api.Character_actor:
		return m.validateActorAction(user, method, info)
	default:
		return status.Error(codes.Internal, "could not find character")
	}
}

func (m *Manager) validateDirectorAction(user *api.User, method api.Method, info interface{}) error {
	switch method {
	case api.Method_GetActorWithID:
		return nil
	case api.Method_GetActorsWithRole:
		return nil
	case api.Method_GetAllActors:
		return nil
	case api.Method_PutResult:
		return nil
	case api.Method_Put:
		key, ok := info.(*api.Key)
		if !ok {
		}
		return m.validateMessageAPIMethods(api.Character_director, "", method)
	case api.Method_Get:
	case api.Method_GetAll:
	case api.Method_Delete:
	case api.Method_DeleteAll:
	case api.Method_Watch:
	default:
		return status.Errorf(codes.PermissionDenied, "invalid action by the director")
	}
}

func (m *Manager) validateMasterAction(user *api.User, method api.Method, info interface{}) error {
	return nil
}

func (m *Manager) validateOperatorAction(user *api.User, method api.Method, info interface{}) error {
	return nil
}

func (m *Manager) validateActorAction(user *api.User, role string, method api.Method, info interface{}) error {

	return status.Errorf(codes.PermissionDenied, "")
}

func (m *Manager) validateMessageAPIMethods(character api.Character, role string, method api.Method, check *api.Key) error {
	keys, err := m.roleStorage.GetRules(character, role, method)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if m.match(base, check) {
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "could not find any rule to match with your action")
}

func (m *Manager) SetupMessage(token string, message *api.Message) error {
	user, err := m.userStorage.GetUserWithToken(token)
	if err != nil {
		return err
	}

	message.Meta = &api.Meta{
		CreationTime: timestamppb.Now(),
		Owner:        user.Id,
		Role:         user.Role,
	}

	return nil
}

func (m *Manager) match(base, check *api.Key) bool {
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