package api

import (
	"context"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/types.go"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	api.UnimplementedAPIServer
	auther    *auth.Auther
	simulator *simulator.Simulator
}

func (s *Server) Get(ctx context.Context, key *api.Key) (*api.Message, error) {
	id, err := s.ExtractIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Validate(id, types.GetMethod, key); err != nil {
		return nil, err
	}

	if err := validateKey(key, types.GetMethod); err != nil {
		return nil, err
	}

	return s.simulator.Get(key)
}

func (s *Server) GetAll(key *api.Key, stream api.API_GetAllServer) error {
	ctx := stream.Context()

	id, err := s.ExtractIDFromContext(ctx)
	if err != nil {
		return err
	}

	if err := s.auther.Validate(id, types.GetAllMethod, key); err != nil {
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
	id, err := s.ExtractIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Validate(id, types.PutMethod, message.Key); err != nil {
		return nil, err
	}

	if err := validateKey(message.Key, types.PutMethod); err != nil {
		return nil, err
	}

	message.Meta = &api.Meta{
		CreationTime: timestamppb.Now(),
		Owner:        id,
	}

	if err := s.simulator.Put(message); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Delete(ctx context.Context, key *api.Key) (*empty.Empty, error) {
	id, err := s.ExtractIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Validate(id, types.DeleteMethod, key); err != nil {
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
	id, err := s.ExtractIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.auther.Validate(id, types.DeleteAllMethod, key); err != nil {
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
	id, err := s.ExtractIDFromContext(stream.Context())
	if err != nil {
		return err
	}

	if err := s.auther.Validate(id, types.WatchMethod, key); err != nil {
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

func (s *Server) ExtractIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, "could not extract metadata from incoming context")
	}

	ids := md.Get("id")
	if len(ids) != 1 {
		return "", status.Errorf(codes.InvalidArgument, "could not extract id from metadata of incoming context")
	}

	return ids[0], nil
}
