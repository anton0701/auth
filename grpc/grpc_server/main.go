package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	config "github.com/anton0701/auth/config"
	env "github.com/anton0701/auth/config/env"
	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
	"github.com/anton0701/auth/internal/repository"
	"github.com/anton0701/auth/internal/repository/auth"
)

const (
	grpcUserAPIDesc = "User-API-v1"
)

type server struct {
	desc.UnimplementedUserV1Server
	dbPool         *pgxpool.Pool
	log            *zap.Logger
	authRepository repository.AuthRepository
}

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	logger, err := initLogger()
	if err != nil {
		log.Fatalf("%s\nUnable to init logger, error: %#v", grpcUserAPIDesc, err)
	}

	err = config.Load(configPath)
	if err != nil {
		logger.Fatal("Unable to load config", zap.Error(err))
	}

	grpcConfig, err := env.NewGRPCConfig()
	if err != nil {
		logger.Fatal("Unable to get grpc config", zap.Error(err))
	}

	pgConfig, err := env.NewPGConfig()
	if err != nil {
		logger.Fatal("Unable to get postgres config", zap.Error(err))
	}

	lis, err := net.Listen("tcp", grpcConfig.Address())
	if err != nil {
		logger.Panic("Failed to listen", zap.Error(err))
	}

	pool, err := pgxpool.Connect(ctx, pgConfig.DSN())
	if err != nil {
		logger.Panic("Unable to connect to db", zap.Error(err))
	}

	authRepo := auth.NewRepository(pool)

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserV1Server(s, &server{dbPool: pool, log: logger, authRepository: authRepo})

	logger.Info("Server listening at", zap.Any("Address", lis.Addr()))

	if err = s.Serve(lis); err != nil {
		logger.Panic("Failed to serve", zap.Error(err))
	}
}

func initLogger() (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("API", grpcUserAPIDesc))
	return logger, nil
}

// GetUserInfo возвращает данные о пользователе на основе запроса.
//
// Запрос включает в себя только ID пользователя.
//
// Параметры:
//   - ctx: контекст для выполнения операции, позволяет отменять или ограничивать по времени выполнение метода.
//   - req: запрос с данными о пользователе.
//
// Возвращает:
//   - *GetUserInfoResponse - структура с данными о пользователе.
//   - error - ошибка, если что-то пошло не так.
func (s *server) GetUserInfo(ctx context.Context, req *desc.GetUserInfoRequest) (*desc.GetUserInfoResponse, error) {
	s.log.Info("Method Get-User", zap.Any("Input params", req))

	// Валидация запроса
	if err := req.Validate(); err != nil {
		s.log.Error("Method Get-User", zap.Error(err))
		return nil, err
	}

	resp, err := s.authRepository.GetUser(ctx, req)

	return resp, err
}

// CreateUser создает нового пользователя.
//
// Запрос содержит данные об имени, email, роли юзера, пароле, повторе пароля (для валидации корректности ввода пароля).
//
// Параметры:
//   - ctx: контекст для выполнения операции.
//   - req: запрос на создание пользователя с данными пользователя.
//
// Возвращает:
//   - *CreateUserResponse: структура с ID созданного пользователя.
//   - error: ошибка, если что-то пошло не так.
func (s *server) CreateUser(ctx context.Context, req *desc.CreateUserRequest) (*desc.CreateUserResponse, error) {
	s.log.Info("Method Create-User", zap.Any("Input params", req))

	// Валидация запроса
	if err := req.Validate(); err != nil {
		s.log.Error("Method Create-User. Invalid input.", zap.Error(err))
		return nil, err
	}

	resp, err := s.authRepository.CreateUser(ctx, req)

	return resp, err
}

// UpdateUser обновляет данные существующего пользователя.
//
// Параметры:
//   - ctx: контекст для выполнения операции.
//   - req: запрос с данными пользователя для обновления.
//
// Возвращает:
//   - *emptypb.Empty - пустая структура, если метод выполнился корректно.
//   - error - ошибка, если что-то пошло не так.
func (s *server) UpdateUser(ctx context.Context, req *desc.UpdateUserRequest) (*emptypb.Empty, error) {
	s.log.Info("Method Update-User", zap.Any("Input params", req))

	// Валидация запроса
	if err := req.Validate(); err != nil {
		s.log.Error("Method Update-User. Invalid input", zap.Error(err))
		return nil, err
	}

	err := s.authRepository.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteUser удаляет данные о существующем пользователе.
//
// Параметры:
//   - ctx: контекст выполнения операции.
//   - req: запрос с данными об удаляемом пользователе (содержит только ID пользователя).
//
// Возвращает:
//   - *emptypb.Empty - пустая структура, если метод выполнился корректно.
//   - error - если что-то пошло не так.
func (s *server) DeleteUser(ctx context.Context, req *desc.DeleteUserRequest) (*emptypb.Empty, error) {
	s.log.Info("Method Delete-User", zap.Any("Input params", req))

	// Валидация запроса
	if err := req.Validate(); err != nil {
		s.log.Error("Method Delete-User. Invalid input", zap.Error(err))
		return nil, err
	}

	err := s.authRepository.DeleteUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
