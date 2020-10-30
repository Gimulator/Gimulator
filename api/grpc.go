package api

import (
	"context"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	api.UnimplementedAPIServer
	um        *auth.UserManager
	simulator *simulator.Simulator
	log       *logrus.Entry
}

func NewServer(um *auth.UserManager, sim *simulator.Simulator) (*Server, error) {
	return &Server{
		um:        um,
		simulator: sim,
		log:       logrus.WithField("component", "api"),
	}, nil
}

func (s *Server) Get(ctx context.Context, key *api.Key) (*api.Message, error) {
	log := s.log.WithField("key", key.String()).WithField("method", "GET")
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.um.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.ID).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.um.Authorize(user.Role, types.GetMethod, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, types.GetMethod); err != nil {
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

func (s *Server) GetAll(key *api.Key, stream api.API_GetAllServer) error {
	log := s.log.WithField("key", key.String()).WithField("method", "GETALL")
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	ctx := stream.Context()
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")

		return err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.um.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("id", user.ID).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.um.Authorize(user.Role, types.GetAllMethod, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, types.GetAllMethod); err != nil {
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
	log := s.log.WithField("key", message.Key.String()).WithField("method", "PUT")
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.um.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.ID).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.um.Authorize(user.Role, types.PutMethod, message.Key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(message.Key, types.PutMethod); err != nil {
		log.WithError(err).Error("could not validate key")
		return nil, err
	}

	log.Info("starting to setup message")
	if err := s.um.SetupMessage(token, message); err != nil {
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
	log := s.log.WithField("key", key.String()).WithField("method", "DELETE")
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.um.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.ID).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.um.Authorize(user.Role, types.DeleteMethod, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, types.DeleteMethod); err != nil {
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
	log := s.log.WithField("key", key.String()).WithField("method", "DELETEALL")
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return nil, err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.um.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return nil, err
	}
	log = log.WithField("id", user.ID).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.um.Authorize(user.Role, types.DeleteAllMethod, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return nil, err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, types.DeleteAllMethod); err != nil {
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

func (s *Server) Watch(key *api.Key, stream api.API_WatchServer) error {
	log := s.log.WithField("key", key.String()).WithField("method", "WATCH")
	log.Info("starting to handle incoming request")

	log.Info("starting to extract token from context")
	ctx := stream.Context()
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("could not extract token form context")
		return err
	}

	log.Info("starting to authenticate incoming request")
	user, err := s.um.Authenticate(token)
	if err != nil {
		log.WithError(err).Error("could not authenticate incoming request")
		return err
	}
	log = log.WithField("id", user.ID).WithField("role", user.Role)

	log.Info("starting to authorize incoming request")
	if err := s.um.Authorize(user.Role, types.WatchMethod, key); err != nil {
		log.WithError(err).Error("could not authorize incoming request")
		return err
	}

	log.Info("starting to validate key")
	if err := validateKey(key, types.WatchMethod); err != nil {
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

	for {
		message := <-send.Ch

		if err := stream.Send(message); err != nil {
			log.WithError(err).Error("could not send message, closing the watch conn")
			return err
		}
	}
}

func (s *Server) ExtractTokenFromContext(ctx context.Context) (string, error) {
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
