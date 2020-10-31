package api

import (
	"context"

	"github.com/Gimulator/Gimulator/manager"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	api.UnimplementedMessageAPIServer
	api.UnimplementedOperatorAPIServer
	api.UnimplementedDirectorAPIServer
	api.UnimplementedActorAPIServer

	manager   *manager.Manager
	simulator *simulator.Simulator
	log       *logrus.Entry
}

func NewServer(manager *manager.Manager, sim *simulator.Simulator) (*Server, error) {
	return &Server{
		manager:   manager,
		simulator: sim,
		log:       logrus.WithField("component", "api"),
	}, nil
}

///////////////////////////////////////////////////////
///////////////////////// MessageAPI Implementation ///
///////////////////////////////////////////////////////
func (s *Server) Get(ctx context.Context, key *api.Key) (*api.Message, error) {
	log := s.log.WithField("key", key.String()).WithField("method", api.Method_Get)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_Get, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, api.Method_Get); err != nil {
		log.WithError(err).Error("could not validate key")
		return nil, err
	}

	log.Info("starting to process incoming request")
	message, err := s.simulator.Get(key)
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return message, nil
}

func (s *Server) GetAll(key *api.Key, stream api.MessageAPI_GetAllServer) error {
	log := s.log.WithField("key", key.String()).WithField("method", api.Method_GetAll)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	ctx := stream.Context()
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")

		return err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_GetAll, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, api.Method_GetAll); err != nil {
		log.WithError(err).Error("could not validate key")
		return err
	}

	log.Info("starting to process incoming request")
	messages, err := s.simulator.GetAll(key)
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return err
	}

	log.Info("starting to send messages")
	for _, mes := range messages {
		if err := stream.Send(mes); err != nil {
			log.WithError(err).Error("could not send message")
			return err
		}
	}

	return nil
}

func (s *Server) Put(ctx context.Context, message *api.Message) (*empty.Empty, error) {
	log := s.log.WithField("key", message.Key.String()).WithField("method", api.Method_Put)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_Put, message.Key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(message.Key, api.Method_Put); err != nil {
		log.WithError(err).Error("could not validate key")
		return nil, err
	}

	log.Info("starting to setup message")
	if err := s.manager.SetupMessage(token, message); err != nil {
		log.WithError(err).Error("could not setup message")
		return nil, err
	}

	log.Info("starting to process incoming request")
	if err := s.simulator.Put(message); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Delete(ctx context.Context, key *api.Key) (*empty.Empty, error) {
	log := s.log.WithField("key", key.String()).WithField("method", api.Method_Delete)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_Delete, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, api.Method_Delete); err != nil {
		log.WithError(err).Error("could not validate key")
		return nil, err
	}

	log.Info("starting to process incoming request")
	if err := s.simulator.Delete(key); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) DeleteAll(ctx context.Context, key *api.Key) (*empty.Empty, error) {
	log := s.log.WithField("key", key.String()).WithField("method", api.Method_DeleteAll)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_DeleteAll, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, api.Method_DeleteAll); err != nil {
		log.WithError(err).Error("could not validate key")
		return nil, err
	}

	log.Info("starting to process incoming request")
	if err := s.simulator.DeleteAll(key); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Watch(key *api.Key, stream api.MessageAPI_WatchServer) error {
	log := s.log.WithField("key", key.String()).WithField("method", api.Method_Watch)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	ctx := stream.Context()
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_Watch, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, api.Method_Watch); err != nil {
		log.WithError(err).Error("could not validate key")
		return err
	}

	send := simulator.NewChannel()
	defer send.Close()

	log.Info("starting to process incoming request")
	if err := s.simulator.Watch(key, send); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return err
	}

	log.Info("starting to send answer of processed request")
	for {
		message := <-send.Ch

		if err := stream.Send(message); err != nil {
			log.WithError(err).Error("could not send answer of processed request, closing the connection...")
			return err
		}
	}
}

///////////////////////////////////////////////////////
//////////////////////// OperatorAPI Implementation ///
///////////////////////////////////////////////////////

func (s *Server) SetUserStatusUnknown(ctx context.Context, id *api.ID) (*empty.Empty, error) {
	log := s.log.WithField("asked-id", id.Id).WithField("method", api.Method_SetUserStatusUnknown)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_SetUserStatusUnknown, id); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to process incoming request")
	if err := s.manager.UpdateStatus(id.Id, api.Status_unknown); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Server) SetUserStatusRunning(ctx context.Context, id *api.ID) (*empty.Empty, error) {
	log := s.log.WithField("asked-id", id.Id).WithField("method", api.Method_SetUserStatusRunning)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_SetUserStatusRunning, id); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to process incoming request")
	if err := s.manager.UpdateStatus(id.Id, api.Status_running); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Server) SetUserStatusFailed(ctx context.Context, id *api.ID) (*empty.Empty, error) {
	log := s.log.WithField("asked-id", id.Id).WithField("method", api.Method_SetUserStatusFailed)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_SetUserStatusFailed, id); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to process incoming request")
	if err := s.manager.UpdateStatus(id.Id, api.Status_failed); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}
	return &empty.Empty{}, nil
}

///////////////////////////////////////////////////////
//////////////////////// DirectorAPI Implementation ///
///////////////////////////////////////////////////////

func (s *Server) GetActorWithID(ctx context.Context, id *api.ID) (*api.Actor, error) {
	log := s.log.WithField("asked-id", id.Id).WithField("method", api.Method_GetActorWithID)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_GetActorWithID, id); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to process incoming request")
	u, err := s.manager.GetUserWithID(id.Id)
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	actor := &api.Actor{
		Id:        u.Id,
		Role:      u.Role,
		Status:    u.Status,
		Readiness: u.Readiness,
	}
	return actor, nil
}

func (s *Server) GetActorsWithRole(role *api.Role, stream api.DirectorAPI_GetActorsWithRoleServer) error {
	log := s.log.WithField("asked-role", role.Role).WithField("method", api.Method_GetActorsWithRole)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	ctx := stream.Context()
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_GetActorWithID, role); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Info("starting to process incoming request")
	users, err := s.manager.GetUsersWithCharacter(api.Character_actor)
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return err
	}

	log.Info("starting to send answer of processed request")
	for _, u := range users {
		if u.Role != role.Role {
			continue
		}

		if err := stream.Send(&api.Actor{
			Id:        u.Id,
			Role:      u.Role,
			Readiness: u.Readiness,
			Status:    u.Status,
		}); err != nil {
			log.WithError(err).Error("could not send answer of processed request, closing the connection...")
			return err
		}
	}

	return nil
}

func (s *Server) GetAllActors(empty *empty.Empty, stream api.DirectorAPI_GetAllActorsServer) error {
	log := s.log.WithField("method", api.Method_GetActorsWithRole)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	ctx := stream.Context()
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_GetActorWithID, nil); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Info("starting to process incoming request")
	users, err := s.manager.GetUsersWithCharacter(api.Character_actor)
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return err
	}

	log.Info("starting to send answer of processed request")
	for _, u := range users {
		if err := stream.Send(&api.Actor{
			Id:        u.Id,
			Role:      u.Role,
			Readiness: u.Readiness,
			Status:    u.Status,
		}); err != nil {
			log.WithError(err).Error("could not send answer of processed request, closing the connection...")
			return err
		}
	}

	return nil
}

func (s *Server) PutResult(ctx context.Context, result *api.Result) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutResult not implemented")
}

///////////////////////////////////////////////////////
/////////////////////////// ActorAPI Implementation ///
///////////////////////////////////////////////////////

func (s *Server) ImReady(ctx context.Context, emp *empty.Empty) (*empty.Empty, error) {
	log := s.log.WithField("method", api.Method_ImReadyMethod)
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.Id).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.manager.Authorize(user, api.Method_ImReadyMethod, nil); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to process incoming request")
	if err := s.manager.UpdateReadiness(user.Id, true); err != nil {
		log.WithError(err).Error("could not process incoming request")
	}

	return &empty.Empty{}, nil
}
