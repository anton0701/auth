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
const grpcUserApiDesc = "User-Api-v1"

type server struct {
	desc.UnimplementedUserV1Server
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	log.Println(grpcUserApiDesc)
	log.Printf("Method Get. Input params:\nId: %d\n************\n\n",
		req.GetId())

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
	log.Println(grpcUserApiDesc)
	log.Printf("Method Create. Input params:\nName: %s\nEmail: %s\nPassword: %s\nPasswordConfirm: %s\nRole: %s\n************\n\n",
		req.GetName(),
		req.GetEmail(),
		req.GetPassword(),
		req.GetPasswordConfirm(),
		req.GetRole())

	resp := &desc.CreateResponse{
		Id: 1,
	}

	return resp, nil
}

func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	log.Println(grpcUserApiDesc)
	log.Printf("Method Update. Input params:\nId: %d\nName: %s\nEmail: %s\nRole: %s\n************\n\n",
		req.GetId(),
		req.GetName(),
		req.GetEmail(),
		req.GetRole())

	resp := &emptypb.Empty{}

	return resp, nil
}

func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	log.Println(grpcUserApiDesc)
	log.Printf("Method Delete. Input params:\nId: %d\n************\n\n",
		req.GetId())

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

	log.Printf("server listening at %v\n\n", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
