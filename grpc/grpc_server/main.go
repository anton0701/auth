package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
)

const (
	grpcPort        = 50051
	grpcUserAPIDesc = "User-API-v1"
	dbDSN           = "host=localhost port=54321 dbname=auth user=auth-user password=auth-password"
)

type server struct {
	desc.UnimplementedUserV1Server
	dbPool *pgxpool.Pool
}

func main() {
	ctx := context.Background()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))

	if err != nil {
		log.Fatalf("%s\nfailed to listen: %v", grpcUserAPIDesc, err)
	}

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatalf("%s\nUnable to connect to db, error: %#v", grpcUserAPIDesc, err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserV1Server(s, &server{dbPool: pool})

	log.Printf("server listening at %v\n\n", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("%s\nfailed to serve: %v", grpcUserAPIDesc, err)
	}
}

func (s *server) GetUserInfo(ctx context.Context, req *desc.GetUserInfoRequest) (*desc.GetUserInfoResponse, error) {
	log.Printf("%s\nMethod Get-User.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	builderSelect := sq.
		Select("id", "name", "email", "role", "created_at", "updated_at").
		From("auth").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.Id})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Fatalf("%s\nMethod Get-User.\nUnable to create sql from builderInsert, error: %#v", grpcUserAPIDesc, err)
	}

	var (
		id          int64
		name, email string
		role        desc.UserRole
		createdAt   time.Time
		updatedAt   sql.NullTime
	)

	err = s.dbPool.
		QueryRow(ctx, query, args...).
		Scan(&id, &name, &email, &role, &createdAt, &updatedAt)

	if err != nil {
		log.Fatalf("%s\nMethod Get-User.\nFatal while query row, error: %#v", grpcUserAPIDesc, err)
	}

	var updatedAtProto *timestamppb.Timestamp
	if updatedAt.Valid {
		updatedAtProto = timestamppb.New(updatedAt.Time)
	} else {
		updatedAtProto = nil
	}

	return &desc.GetUserInfoResponse{
		Id:        id,
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: updatedAtProto,
	}, nil
}

func (s *server) CreateUser(ctx context.Context, req *desc.CreateUserRequest) (*desc.CreateUserResponse, error) {
	log.Printf("%s\nMethod Create-User.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	builderInsert := sq.Insert("auth").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "email", "password", "password_confirm", "role").
		Values(req.Name, req.Email, req.Password, req.PasswordConfirm, int32(req.Role)).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("%s\nMethod Create-User.\nUnable to create sql from builderInsert, error: %#v", grpcUserAPIDesc, err)
	}
	log.Printf("Generated SQL Query: %s", query)
	log.Printf("Arguments: %v", args)

	var userID int64
	err = s.dbPool.
		QueryRow(ctx, query, args...).
		Scan(&userID)

	if err != nil {
		log.Fatalf("%s\nMethod Create-User.\nUnable to get userID from created user, error: %#v", grpcUserAPIDesc, err)
	}

	return &desc.CreateUserResponse{
		Id: userID,
	}, nil
}

func (s *server) UpdateUser(ctx context.Context, req *desc.UpdateUserRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod Update.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	builderUpdate := sq.
		Update("auth").
		PlaceholderFormat(sq.Dollar).
		Set("name", req.Name).
		Set("email", req.Email).
		Set("role", int32(req.GetRole())).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.Id})

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		//log.Fatalf("%s\nMethod Update-User.\nUnable to create sql from builderUpdate, error: %#v", grpcUserAPIDesc, err)
		log.Printf("%s\nMethod Update-User.\nUnable to create sql from builderUpdate, error: %#v", grpcUserAPIDesc, err)
	}

	_, err = s.dbPool.Exec(ctx, query, args...)
	if err != nil {
		//log.Fatalf("%s\nMethod Update-User.\nUnable to execute sql query, error: %#v", grpcUserAPIDesc, err)
		log.Printf("%s\nMethod Update-User.\nUnable to execute sql query, error: %#v", grpcUserAPIDesc, err)
	}

	return &emptypb.Empty{}, nil
}

func (s *server) DeleteUser(ctx context.Context, req *desc.DeleteUserRequest) (*emptypb.Empty, error) {
	log.Printf("%s\nMethod Delete.\nInput params:\n%+v\n************\n\n", grpcUserAPIDesc, req)

	builderDelete := sq.Delete("auth").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.Id})

	query, args, err := builderDelete.ToSql()
	if err != nil {
		log.Printf("%s\nMethod Delete-User.\nUnable to create sql from builderDelete, error: %#v", grpcUserAPIDesc, err)
	}

	_, err = s.dbPool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("%s\nMethod Delete-User.\nUnable to execute sql query, error: %#v", grpcUserAPIDesc, err)
	}

	return &emptypb.Empty{}, nil
}
