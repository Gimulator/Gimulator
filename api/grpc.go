package api

import (
	"context"
	"os"
	"time"
	"sync"

	"github.com/Gimulator/Gimulator/manager"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	api.UnimplementedMessageAPIServer
	api.UnimplementedOperatorAPIServer
	api.UnimplementedDirectorAPIServer
	api.UnimplementedUserAPIServer

	manager   *manager.Manager
	simulator *simulator.Simulator
	log       *logrus.Entry

	mutex     sync.Mutex
}

func NewServer(manager *manager.Manager, sim *simulator.Simulator) (*Server, error) {
	return &Server{
		manager:   manager,
		simulator: sim,
		log:       logrus.WithField("component", "grpc"),
	}, nil
}

func (s *Server) finalizeGame(result *api.Result) {
	s.log.Debug("starting to process incoming request")
	for {
		err := s.manager.Epilogue.Write(result)
		if err == nil {
			break
		}
		s.log.WithError(err).Error("could not write result into epilogue")
		time.Sleep(5 * time.Second)
	}

	// TODO Close the gRPC server gracefully

	// Shutdown Gimulator
	os.Exit(0)
}

///////////////////////////////////////////////////////
///////////////////////// MessageAPI Implementation ///
///////////////////////////////////////////////////////

func (s *Server) Get(ctx context.Context, key *api.Key) (*api.Message, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("key", key.String()).WithField("method", api.Method_get)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeGetMethod(user, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Debug("starting to process incoming request")
	message, err := s.simulator.Get(key)
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return message, nil
}

func (s *Server) GetAll(key *api.Key, stream api.MessageAPI_GetAllServer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("key", key.String()).WithField("method", api.Method_getAll)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	ctx := stream.Context()
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")

		return err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeGetAllMethod(user, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Debug("starting to process incoming request")
	messages, err := s.simulator.GetAll(key)
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return err
	}

	log.Debug("starting to send messages")
	for _, mes := range messages {
		if err := stream.Send(mes); err != nil {
			log.WithError(err).Error("could not send message")
			return err
		}
	}

	return nil
}

func (s *Server) Put(ctx context.Context, message *api.Message) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("key", message.Key.String()).WithField("method", api.Method_put)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizePutMethod(user, message.Key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Debug("starting to setup the meta of the message")
	message.Meta = &api.Meta{
		Owner: &api.User{
			Name:      user.Name,
			Character: user.Character,
			Role:      user.Role,
			Readiness: user.Readiness,
			Status:    user.Status,
		},
	}

	log.Debug("starting to process incoming request")
	if err := s.simulator.Put(message); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Delete(ctx context.Context, key *api.Key) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("key", key.String()).WithField("method", api.Method_delete)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeDeleteMethod(user, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Debug("starting to process incoming request")
	if err := s.simulator.Delete(key); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) DeleteAll(ctx context.Context, key *api.Key) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("key", key.String()).WithField("method", api.Method_deleteAll)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeDeleteAllMethod(user, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Debug("starting to process incoming request")
	if err := s.simulator.DeleteAll(key); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Watch(key *api.Key, stream api.MessageAPI_WatchServer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("key", key.String()).WithField("method", api.Method_watch)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	ctx := stream.Context()
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeWatchMethod(user, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	send := simulator.NewChannel()
	defer send.Close()

	log.Debug("starting to process incoming request")
	if err := s.simulator.Watch(key, send); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return err
	}

	log.Debug("starting to send answer of processed request")
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

func (s *Server) SetUserStatus(ctx context.Context, report *api.Report) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("asked-name", report.Name).WithField("status", report.Status).WithField("method", api.Method_setUserStatus)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeSetUserStatusMethod(user, report); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Debug("starting to process incoming request")
	if err := s.manager.UpdateStatus(report.Name, report.Status); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}
	return &empty.Empty{}, nil
}

///////////////////////////////////////////////////////
//////////////////////// DirectorAPI Implementation ///
///////////////////////////////////////////////////////

func (s *Server) GetActors(empty *empty.Empty, stream api.DirectorAPI_GetActorsServer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("method", api.Method_getActors)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	ctx := stream.Context()
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeGetActorsMethod(user); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Debug("starting to process incoming request")
	users, err := s.manager.GetActors()
	if err != nil {
		log.WithError(err).Error("could not process incoming request")
		return err
	}

	log.Debug("starting to send answer of processed request")
	for _, u := range users {
		if err := stream.Send(u); err != nil {
			log.WithError(err).Error("could not send answer of processed request, closing the connection...")
			return err
		}
	}

	return nil
}

func (s *Server) PutResult(ctx context.Context, result *api.Result) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("method", api.Method_putResult)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizePutResultMethod(user); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	go s.finalizeGame(result)

	return &empty.Empty{}, nil
}

///////////////////////////////////////////////////////
//////////////////////////// UserAPI Implementation ///
///////////////////////////////////////////////////////

func (s *Server) ImReady(ctx context.Context, emp *empty.Empty) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("method", api.Method_imReady)
	log.Debug("starting to handle incoming request")

	log.Debug("starting to extract token from context")
	token, err := s.extractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Debug("starting to authenticate incoming request")
	user, err := s.manager.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("name", user.Name).WithField("role", user.Role)

	log.Debug("starting to authorize incoming request")
	if err := s.manager.AuthorizeImReadyMethod(user); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Debug("starting to process incoming request")
	if err := s.manager.UpdateReadiness(user.Name, true); err != nil {
		log.WithError(err).Error("could not process incoming request")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Ping(ctx context.Context, emp *empty.Empty) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := s.log.WithField("method", api.Method_ping)
	log.Debug("starting to handle incoming request")

	return &empty.Empty{}, nil
}

///////////////////////////////////////////////////////
//////////////////////////////////////////// Helper ///
///////////////////////////////////////////////////////

func (s *Server) extractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, "could not extract metadata from incoming context")
	}

	tokens := md.Get("token")
	if len(tokens) != 1 {
		return "", status.Errorf(codes.InvalidArgument, "could not extract token from metadata of incoming context")
	}

	return tokens[0], nil
}
