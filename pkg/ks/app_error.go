package ks

import (
	"fmt"
)

type AppError struct {
	Code    string
	Message string
	Cause   error
}

func (my *AppError) Error() string {
	if my.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", my.Code, my.Message, my.Cause)
	}
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

func WrapAppError(code, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
