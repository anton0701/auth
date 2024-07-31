package auth

import (
	"context"
	"database/sql"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
	"github.com/anton0701/auth/internal/repository"
)

const (
	tableName = "auth"

	idColumn              = "id"
	nameColumn            = "name"
	emailColumn           = "email"
	roleColumn            = "role"
	createdAtColumn       = "created_at"
	updatedAtColumn       = "updated_at"
	passwordColumn        = "password"
	passwordConfirmColumn = "password_confirm"
)

// TODO: какие ошибки возвращать?
type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) repository.AuthRepository {
	return &repo{db: db}
}

func (r *repo) GetUser(ctx context.Context, req *desc.GetUserInfoRequest) (*desc.GetUserInfoResponse, error) {
	builderSelect := sq.
		Select(idColumn, nameColumn, emailColumn, roleColumn, createdAtColumn, updatedAtColumn).
		From(tableName).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{idColumn: req.Id})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		//s.log.Error("Method Get-User. Unable to create SQL query from builder", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to create SQL query from builder. Error info: %v", err)
	}

	var (
		id          int64
		name, email string
		role        desc.UserRole
		createdAt   time.Time
		updatedAt   sql.NullTime
	)

	err = r.db.
		QueryRow(ctx, query, args...).
		Scan(&id, &name, &email, &role, &createdAt, &updatedAt)
	if err != nil {
		//s.log.Error("Method Get-User. Error while query row", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Error while query row. Error info: %v", err)
	}

	var updatedAtProto *timestamppb.Timestamp
	if updatedAt.Valid {
		updatedAtProto = timestamppb.New(updatedAt.Time)
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

func (r *repo) CreateUser(ctx context.Context, req *desc.CreateUserRequest) (*desc.CreateUserResponse, error) {
	builderInsert := sq.Insert(tableName).
		PlaceholderFormat(sq.Dollar).
		Columns(nameColumn, emailColumn, passwordColumn, passwordConfirmColumn, roleColumn).
		Values(req.Name, req.Email, req.Password, req.PasswordConfirm, int32(req.Role)).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		//s.log.Error("Method Create-User. Unable to create SQL query from builder", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to create SQL query from builder, error: %#v", err)
	}

	var userID int64
	err = r.db.
		QueryRow(ctx, query, args...).
		Scan(&userID)
	if err != nil {
		//s.log.Error("Method Create-User. Unable to get userID from created user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Unable to get userID from created user, error: %#v", err)
	}

	return &desc.CreateUserResponse{
		Id: userID,
	}, nil
}

func (r *repo) UpdateUser(ctx context.Context, req *desc.UpdateUserRequest) error {
	builderUpdate := sq.
		Update(tableName).
		PlaceholderFormat(sq.Dollar).
		Set(roleColumn, int32(req.GetRole())).
		Set(updatedAtColumn, time.Now()).
		Where(sq.Eq{idColumn: req.Id})

	if req.Name != nil {
		trimmedName := strings.TrimSpace(req.Name.GetValue())
		if len(trimmedName) > 0 {
			builderUpdate.Set(nameColumn, trimmedName)
		}
	}

	if req.Email != nil {
		trimmedEmail := strings.TrimSpace(req.Email.GetValue())
		if len(trimmedEmail) > 0 {
			builderUpdate.Set(emailColumn, trimmedEmail)
		}
	}

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		//s.log.Error("Method Update-User. Unable to create SQL query from builder", zap.Error(err))
		return status.Errorf(codes.Internal, "Unable to create SQL query from builder, error info: %#v", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		//s.log.Error("Method Update-User. Unable to execute SQL query", zap.Error(err))
		return status.Errorf(codes.Internal, "Unable to execute SQL query, error info: %#v", err)
	}

	return nil
}

func (r *repo) DeleteUser(ctx context.Context, req *desc.DeleteUserRequest) error {
	builderDelete := sq.Delete(tableName).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{idColumn: req.Id})

	query, args, err := builderDelete.ToSql()
	if err != nil {
		//s.log.Error("Method Delete-User. Unable to create SQL query from builder", zap.Error(err))
		return status.Errorf(codes.Internal, "Unable to create SQL query from builder, error info: %#v", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		//s.log.Error("Method Delete-User. Unable to execute SQL query", zap.Error(err))
		return status.Errorf(codes.Internal, "Unable to execute SQL query, error info: %#v", err)
	}

	return nil
}
