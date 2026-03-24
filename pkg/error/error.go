package error

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    string
	Message string
}

func (my *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", my.Code, my.Message)
}

func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewAppErrorf(code, format string, args ...any) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func AsAppError(err error, target **AppError) bool {
	return errors.As(err, target)
}

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}
