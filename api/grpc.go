package api

import (
	"context"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	api.UnimplementedAPIServer
	auther    *auth.Auther
	simulator *simulator.Simulator
}

func NewServer(auther *auth.Auther, sim *simulator.Simulator) (*Server, error) {
	return &Server{
		auther:    auther,
		simulator: sim,
	}, nil
}

func (s *Server) Get(ctx context.Context, key *api.Key) (*api.Message, error) {
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Auth(token, types.GetMethod, key); err != nil {
		return nil, err
	}

	if err := validateKey(key, types.GetMethod); err != nil {
		return nil, err
	}

	return s.simulator.Get(key)
}

func (s *Server) GetAll(key *api.Key, stream api.API_GetAllServer) error {
	ctx := stream.Context()

	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		return err
	}

	if err := s.auther.Auth(token, types.GetAllMethod, key); err != nil {
		return err
	}

	if err := validateKey(key, types.GetAllMethod); err != nil {
		return err
	}

	messages, err := s.simulator.GetAll(key)
	if err != nil {
		return err
	}

	for _, mes := range messages {
		if err := stream.Send(mes); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Put(ctx context.Context, message *api.Message) (*empty.Empty, error) {
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Auth(token, types.PutMethod, message.Key); err != nil {
		return nil, err
	}

	if err := validateKey(message.Key, types.PutMethod); err != nil {
		return nil, err
	}

	s.auther.SetupMessage(token, message)

	if err := s.simulator.Put(message); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Delete(ctx context.Context, key *api.Key) (*empty.Empty, error) {
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Auth(token, types.DeleteMethod, key); err != nil {
		return nil, err
	}

	if err := validateKey(key, types.DeleteMethod); err != nil {
		return nil, err
	}

	if err := s.simulator.Delete(key); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) DeleteAll(ctx context.Context, key *api.Key) (*empty.Empty, error) {
	token, err := s.ExtractTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Auth(token, types.DeleteAllMethod, key); err != nil {
		return nil, err
	}

	if err := validateKey(key, types.DeleteAllMethod); err != nil {
		return nil, err
	}

	if err := s.simulator.Delete(key); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Watch(key *api.Key, stream api.API_WatchServer) error {
	token, err := s.ExtractTokenFromContext(stream.Context())
	if err != nil {
		return err
	}

	if err := s.auther.Auth(token, types.WatchMethod, key); err != nil {
		return err
	}

	if err := validateKey(key, types.WatchMethod); err != nil {
		return err
	}

	send := &simulator.Channel{
		Ch:       make(chan *api.Message),
		IsClosed: false,
	}
	defer func() {
		send.IsClosed = true
		close(send.Ch)
	}()

	if err := s.simulator.Watch(key, send); err != nil {
		return err
	}

	for {
		message := <-send.Ch

		if err := stream.Send(message); err != nil {
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
