package ks

import (
	"fmt"
)

type AppError struct {
	Code    string
	Message string
}

func (my *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", my.Code, my.Message)
}

func NewAppError(code, message string, args ...any) *AppError {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	return &AppError{
		Code:    code,
		Message: message,
	}
}
