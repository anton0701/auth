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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
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
	log    *zap.Logger
}

func main() {
	ctx := context.Background()

	logger, err := initLogger()
	if err != nil {
		log.Fatalf("%s\nUnable to init logger, error: %#v", grpcUserAPIDesc, err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))

	if err != nil {
		logger.Panic("Failed to listen", zap.Error(err))
	}

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		logger.Panic("Unable to connect to db", zap.Error(err))
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserV1Server(s, &server{dbPool: pool, log: logger})

	logger.Info("Server listening at", zap.Any("Address", lis.Addr()))

	if err = s.Serve(lis); err != nil {
		logger.Panic("Failed to serve", zap.Error(err))
	}
}

func initLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("API", grpcUserAPIDesc))
	return logger, nil
}

func (s *server) GetUserInfo(ctx context.Context, req *desc.GetUserInfoRequest) (*desc.GetUserInfoResponse, error) {
	s.log.Info("Method Get-User", zap.Any("Input params", req))

	builderSelect := sq.
		Select("id", "name", "email", "role", "created_at", "updated_at").
		From("auth").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.Id})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		s.log.Error("Method Get-User. Unable to create SQL query from builder", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to create SQL query from builder. Error info: %v", err)
	}
	s.log.Info("Method Get-User. Generated SQL Query",
		zap.String("query", query),
		zap.Any("args", args))

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
		s.log.Error("Method Get-User. Error while query row", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Error while query row. Error info: %v", err)
	}

	// TODO: перепроверить
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
	s.log.Info("Method Create-User", zap.Any("Input params", req))

	builderInsert := sq.Insert("auth").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "email", "password", "password_confirm", "role").
		Values(req.Name, req.Email, req.Password, req.PasswordConfirm, int32(req.Role)).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		s.log.Error("Method Create-User. Unable to create SQL query from builder", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to create SQL query from builder, error: %#v", err)
	}

	var userID int64
	err = s.dbPool.
		QueryRow(ctx, query, args...).
		Scan(&userID)

	if err != nil {
		s.log.Error("Method Create-User. Unable to get userID from created user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to get userID from created user, error: %#v", err)
	}

	return &desc.CreateUserResponse{
		Id: userID,
	}, nil
}

func (s *server) UpdateUser(ctx context.Context, req *desc.UpdateUserRequest) (*emptypb.Empty, error) {
	s.log.Info("Method Update-User", zap.Any("Input params", req))

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
		s.log.Error("Method Update-User. Unable to create SQL query from builder", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to create SQL query from builder, error info: %#v", err)
	}

	_, err = s.dbPool.Exec(ctx, query, args...)
	if err != nil {
		s.log.Error("Method Update-User. Unable to execute SQL query", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to execute SQL query, error info: %#v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *server) DeleteUser(ctx context.Context, req *desc.DeleteUserRequest) (*emptypb.Empty, error) {
	s.log.Info("Method Delete-User", zap.Any("Input params", req))

	builderDelete := sq.Delete("auth").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.Id})

	query, args, err := builderDelete.ToSql()
	if err != nil {
		s.log.Error("Method Delete-User. Unable to create SQL query from builder", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to create SQL query from builder, error info: %#v", err)
	}

	_, err = s.dbPool.Exec(ctx, query, args...)
	if err != nil {
		s.log.Error("Method Delete-User. Unable to execute SQL query", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to execute SQL query, error info: %#v", err)
	}

	return &emptypb.Empty{}, nil
}
