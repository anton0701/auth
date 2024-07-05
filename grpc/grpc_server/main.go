package main

import (
	"context"
	"fmt"
	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"time"
)

const grpcPort = 50051

type server struct {
	desc.UnimplementedUserV1Server
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	fmt.Printf("Method Get. Input params:\nId: %d\n************\n\n", req.Id)
	response := &desc.GetResponse{
		Id:        req.GetId(),
		Name:      "Test Name",
		Email:     "Test Email",
		Role:      desc.UserRole_user,
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}
	return response, nil
}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	fmt.Printf("Method Create. Input params:\nName: %s\nEmail: %s\nPassword: %s\nPasswordConfirm: %s\nRole: %s\n************\n\n", req.Name, req.Email, req.Password, req.PasswordConfirm, req.Role)
	resp := &desc.CreateResponse{
		Id: 1,
	}
	return resp, nil
}

func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	fmt.Printf("Method Update. Input params:\nId: %d\nName: %s\nEmail: %s\nRole: %s\n************\n\n", req.Id, req.Name, req.Email, req.Role)
	resp := &emptypb.Empty{}
	return resp, nil
}

func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	fmt.Printf("Method Delete. Input params:\nId: %d\n************\n\n", req.Id)
	resp := &emptypb.Empty{}
	return resp, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
