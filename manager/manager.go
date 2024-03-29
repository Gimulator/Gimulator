package manager

import (
	"fmt"
	"os"

	"github.com/Gimulator/Gimulator/cmd"
	"github.com/Gimulator/Gimulator/epilogues"
	"github.com/Gimulator/Gimulator/storage"
	"github.com/Gimulator/protobuf/go/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Manager struct {
	userStorage storage.UserStorage
	ruleStorage storage.RuleStorage

	Epilogue epilogues.Epilogue
}

func NewManager(credStorage storage.UserStorage, roleStorage storage.RuleStorage, epilogue epilogues.Epilogue) (*Manager, error) {
	return &Manager{
		userStorage: credStorage,
		ruleStorage: roleStorage,
		Epilogue:    epilogue,
	}, nil
}

func (m *Manager) Authenticate(token string) (*api.User, error) {
	user, err := m.userStorage.GetUserWithToken(token)
	if err != nil {
		return nil, err
	}
	return user, err
}

func (m *Manager) UpdateStatus(name string, status api.Status) error {
	err := m.userStorage.UpdateUserStatus(name, status)

	// Checking if director has failed
	if status == api.Status_failed {
		_user, err2 := m.GetUserWithName(name)
		if err2 != nil {
			return err2
		}
		if _user.GetCharacter() == api.Character_director {
			// Director has failed, so Gimulator must be terminated.

			// Sending report to message queue
			result := api.Result{
				Id:     cmd.Id,
				Msg:    "Director Failed.",
				Status: api.Result_failed,
			}
			err3 := m.Epilogue.Write(&result)
			if err3 != nil {
				return err3
			}

			// Shutting down Gimulator
			os.Exit(0)
		}
	}
	return err
}

func (m *Manager) UpdateReadiness(name string, readiness bool) error {
	return m.userStorage.UpdateUserReadiness(name, readiness)
}

func (m *Manager) GetUserWithName(id string) (*api.User, error) {
	return m.userStorage.GetUserWithName(id)
}

func (m *Manager) GetActors() ([]*api.User, error) {
	char := api.Character_actor
	return m.userStorage.GetUsers(nil, nil, &char, nil, nil, nil)
}

func (m *Manager) AuthorizeGetMethod(user *api.User, key *api.Key) error {
	if err := m.checkKeyNilness(key); err != nil {
		return err
	}

	if err := m.checkKeyCompleteness(key); err != nil {
		return err
	}

	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_get)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if m.match(base, key) {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, fmt.Sprintf("invalid action: you don't have permission to get a message with key=%v", key))
}

func (m *Manager) AuthorizeGetAllMethod(user *api.User, key *api.Key) error {
	if err := m.checkKeyNilness(key); err != nil {
		return err
	}

	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_getAll)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if m.match(base, key) {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, fmt.Sprintf("invalid action: you don't have permission to get all messages with key=%v", key))
}

func (m *Manager) AuthorizePutMethod(user *api.User, key *api.Key) error {
	if err := m.checkKeyNilness(key); err != nil {
		return err
	}

	if err := m.checkKeyCompleteness(key); err != nil {
		return err
	}

	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_put)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if m.match(base, key) {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, fmt.Sprintf("invalid action: you don't have permission to put a message with key=%v", key))
}

func (m *Manager) AuthorizeDeleteMethod(user *api.User, key *api.Key) error {
	if err := m.checkKeyNilness(key); err != nil {
		return err
	}

	if err := m.checkKeyCompleteness(key); err != nil {
		return err
	}

	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_delete)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if m.match(base, key) {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, fmt.Sprintf("invalid action: you don't have permission to delete a message with key=%v", key))
}

func (m *Manager) AuthorizeDeleteAllMethod(user *api.User, key *api.Key) error {
	if err := m.checkKeyNilness(key); err != nil {
		return err
	}

	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_deleteAll)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if m.match(base, key) {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, fmt.Sprintf("invalid action: you don't have permission to delete all messages with key=%v", key))
}

func (m *Manager) AuthorizeWatchMethod(user *api.User, key *api.Key) error {
	if err := m.checkKeyNilness(key); err != nil {
		return err
	}

	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_watch)
	if err != nil {
		return err
	}

	for _, base := range keys {
		if m.match(base, key) {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, fmt.Sprintf("invalid action: you don't have permission to watch messages with key=%v", key))
}

func (m *Manager) AuthorizeSetUserStatusMethod(user *api.User, report *api.Report) error {
	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_setUserStatus)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return nil
	}

	return status.Error(codes.PermissionDenied, "invalid action: you don't have permission to update user's status")
}

func (m *Manager) AuthorizeGetActorsMethod(user *api.User) error {
	fmt.Println(user)
	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_getActors)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return nil
	}

	return status.Error(codes.PermissionDenied, "invalid action: you don't have permission to get actors")
}

func (m *Manager) AuthorizePutResultMethod(user *api.User) error {
	keys, err := m.ruleStorage.GetRules(user.Character, user.Role, api.Method_putResult)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return nil
	}

	return status.Error(codes.PermissionDenied, "invalid action: you don't have permission to put the result of room")
}

func (m *Manager) AuthorizeImReadyMethod(user *api.User) error {
	return nil
}

func (m *Manager) validateMessageAPIMethods(character api.Character, role string, method api.Method, check *api.Key) error {
	keys, err := m.ruleStorage.GetRules(character, role, method)
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

func (m *Manager) checkKeyNilness(k *api.Key) error {
	if k == nil {
		return status.Error(codes.InvalidArgument, "invalid key: key can not be null")
	}
	return nil
}

func (m *Manager) checkKeyCompleteness(k *api.Key) error {
	if k.Type == "" {
		return status.Error(codes.InvalidArgument, "invalid key: type of key can not be empty")
	}
	if k.Name == "" {
		return status.Error(codes.InvalidArgument, "invalid key: name of key can not be empty")
	}
	if k.Namespace == "" {
		return status.Error(codes.InvalidArgument, "invalid key: namespace of key can not be empty")
	}
	return nil
}
