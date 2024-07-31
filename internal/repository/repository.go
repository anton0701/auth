package repository

import (
	"context"
	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
)

type AuthRepository interface {
	GetUser(ctx context.Context, req *desc.GetUserInfoRequest) (*desc.GetUserInfoResponse, error)
	CreateUser(ctx context.Context, req *desc.CreateUserRequest) (*desc.CreateUserResponse, error)
	UpdateUser(ctx context.Context, req *desc.UpdateUserRequest) error
	DeleteUser(ctx context.Context, req *desc.DeleteUserRequest) error
}
