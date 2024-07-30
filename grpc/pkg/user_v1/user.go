package user_v1

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/anton0701/auth/grpc/pkg"
)

var (
	_ pkg.Validator = (*GetUserInfoRequest)(nil)
	_ pkg.Validator = (*CreateUserRequest)(nil)
	_ pkg.Validator = (*UpdateUserRequest)(nil)
	_ pkg.Validator = (*DeleteUserRequest)(nil)
)

// Validate
//
// Возвращает:
//   - error, если User_id не указан.
//   - nil в остальных случаях.
func (req *GetUserInfoRequest) Validate() error {
	// TODO: текст ошибки вынести в константу. Сделаю в рамках ДЗ №3 - слоистая архитектура
	// В запросе должен быть ID (User_ID)
	if req.Id == 0 {
		err := status.Error(codes.InvalidArgument, "User-id must be provided")
		return err
	}

	return nil
}

// Validate
//
// Возвращает:
//   - error, если User_name пустой.
//   - error, если Email пустой.
//   - error, если Password пустой либо не совпадает с Password_confirm.
//   - error, если Role некорректная.
//   - nil в остальных случаях.
func (req *CreateUserRequest) Validate() error {
	// Проверка, что User_name не пустой
	trimmedNameFromRequest := strings.TrimSpace(req.Name)
	if len(trimmedNameFromRequest) == 0 {
		err := status.Error(codes.InvalidArgument, "User name must not be empty")
		return err
	}

	// Проверка, что Email не пустой
	trimmedEmailFromRequest := strings.TrimSpace(req.Email)
	if len(trimmedEmailFromRequest) == 0 {
		err := status.Error(codes.InvalidArgument, "Email must not be empty")
		return err
	}

	// Проверка, что Password не пустой и совпадает с Password_confirm
	trimmedPasswordFromRequest := strings.TrimSpace(req.Password)
	if (req.Password != req.PasswordConfirm) || len(trimmedPasswordFromRequest) == 0 {
		err := status.Error(codes.InvalidArgument, "Password must not be empty. Password must be equal to Password_confirm")
		return err
	}

	// Проверка, что Role корректная
	if req.GetRole() == UserRole_UNKNOWN {
		err := status.Error(codes.InvalidArgument, "Invalid role")
		return err
	}

	return nil
}

// Validate
//
// Возвращает:
//   - error, если Role == UNKNOWN.
//   - nil в остальных случаях.
func (req *UpdateUserRequest) Validate() error {
	// Проверка, что Role корректная
	if req.GetRole() == UserRole_UNKNOWN {
		err := status.Error(codes.InvalidArgument, "Invalid role")
		return err
	}

	return nil
}

// Validate
//
// Возвращает:
//   - error, если User-id не указан.
//   - nil в остальных случаях.
func (req *DeleteUserRequest) Validate() error {
	// Проверка, что User_id указан
	if req.Id == 0 {
		err := status.Error(codes.InvalidArgument, "User-id must be provided")
		return err
	}

	return nil
}
