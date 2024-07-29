package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net"
	"strings"
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

	config "github.com/anton0701/auth/config"
	env "github.com/anton0701/auth/config/env"
	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
)

const (
	grpcUserAPIDesc = "User-API-v1"
)

type server struct {
	desc.UnimplementedUserV1Server
	dbPool *pgxpool.Pool
	log    *zap.Logger
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

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserV1Server(s, &server{dbPool: pool, log: logger})

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

	// Проверка необходимых полей, в запросе должен быть ID (User_ID)
	if req.Id == 0 {
		err := status.Error(codes.InvalidArgument, "User-id must be provided")
		s.log.Error("Method Get-User", zap.Error(err))
		return nil, err
	}

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

	// TODO: правильно?
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

	// TODO: текст ошибки вынести в константу?

	// Проверка наличия и корректности полей запроса
	trimmedNameFromRequest := strings.TrimSpace(req.Name)
	if len(trimmedNameFromRequest) == 0 {
		err := status.Error(codes.InvalidArgument, "User name must not be empty")
		s.log.Error("Method Create-User. Invalid input", zap.Error(err))
		return nil, err
	}

	trimmedEmailFromRequest := strings.TrimSpace(req.Email)
	if len(trimmedEmailFromRequest) == 0 {
		err := status.Error(codes.InvalidArgument, "Email must not be empty")
		s.log.Error("Method Create-User. Invalid input", zap.Error(err))
		return nil, err
	}

	trimmedPasswordFromRequest := strings.TrimSpace(req.Password)
	if (req.Password != req.PasswordConfirm) || len(trimmedPasswordFromRequest) == 0 {
		err := status.Error(codes.InvalidArgument, "Password must not be empty. Password must be equal to Password_confirm")
		s.log.Error("Method Create-User. Invalid input", zap.Error(err))
		return nil, err
	}

	if req.GetRole() == desc.UserRole_UNKNOWN {
		err := status.Error(codes.InvalidArgument, "Invalid role")
		s.log.Error("Method Create-User. Invalid input", zap.Error(err))
		return nil, err
	}

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

	// Проверка наличия и корректности полей запроса
	if req.GetRole() == desc.UserRole_UNKNOWN {
		err := status.Error(codes.InvalidArgument, "Invalid role")
		s.log.Error("Method Update-User. Invalid input", zap.Error(err))
		return nil, err
	}

	builderUpdate := sq.
		Update("auth").
		PlaceholderFormat(sq.Dollar).
		Set("role", int32(req.GetRole())).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.Id})

	if req.Name != nil {
		trimmedName := strings.TrimSpace(req.Name.GetValue())
		if len(trimmedName) > 0 {
			builderUpdate.Set("name", trimmedName)
		}
	}

	if req.Email != nil {
		trimmedEmail := strings.TrimSpace(req.Email.GetValue())
		if len(trimmedEmail) > 0 {
			builderUpdate.Set("name", trimmedEmail)
		}
	}

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

	if req.Id == 0 {
		err := status.Error(codes.InvalidArgument, "User-id must be provided")
		s.log.Error("Method Delete-User", zap.Error(err))
		return nil, err
	}

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
