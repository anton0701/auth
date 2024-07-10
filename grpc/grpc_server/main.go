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
	grpcUserApiDesc = "User-Api-v1"
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

func (s *server) Get(_ context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	log.Printf("%s\nMethod Get.\nInput params:\n%+v\n************\n\n", grpcUserApiDesc, req)

	return &desc.GetResponse{
		Id:        req.GetId(),
		Name:      "Test Name",
		Email:     "Test Email",
		Role:      desc.UserRole_USER,
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}, nil
}

func (s *server) Create(_ context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	log.Printf("%s\nMethod Create.\nInput params:\n%+v\n************\n\n", grpcUserApiDesc, req)

	return &desc.CreateResponse{
		Id: 1,
	}, nil
}

func (s *server) Update(_ context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod Update.\nInput params:\n%+v\n************\n\n", grpcUserApiDesc, req)

	return &emptypb.Empty{}, nil
}

func (s *server) Delete(_ context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod Delete.\nInput params:\n%+v\n************\n\n", grpcUserApiDesc, req)

	return &emptypb.Empty{}, nil
}
