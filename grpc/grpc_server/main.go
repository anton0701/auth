package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
)

const (
	grpcPort        = 50051
	grpcUserAPIDesc = "User-API-v1"
)

type server struct {
	desc.UnimplementedUserV1Server
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserV1Server(s, &server{})

	log.Printf("server listening at %v\n\n", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) GetUserInfo(_ context.Context, req *desc.GetUserInfoRequest) (*desc.GetUserInfoResponse, error) {
	log.Printf("%s\nMethod Get.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	return &desc.GetUserInfoResponse{
		Id:        req.GetId(),
		Name:      "Test Name",
		Email:     "test@email.com",
		Role:      desc.UserRole_USER,
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}, nil
}

func (s *server) CreateUser(_ context.Context, req *desc.CreateUserRequest) (*desc.CreateUserResponse, error) {
	log.Printf("%s\nMethod Create.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	return &desc.CreateUserResponse{
		Id: 1,
	}, nil
}

func (s *server) UpdateUser(_ context.Context, req *desc.UpdateUserRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod Update.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	return &emptypb.Empty{}, nil
}

func (s *server) DeleteUser(_ context.Context, req *desc.DeleteUserRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod Delete.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	return &emptypb.Empty{}, nil
}
