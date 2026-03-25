package ks

import (
	"fmt"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/logo/tools"
)

type AppError struct {
	Code    string
	Message string
}

func (my *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", my.Code, my.Message)
}

func NewError(code, message string, args ...any) *AppError {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	return &AppError{
		Code:    code,
		Message: message,
	}
}

func TraceError(code string, args ...any) *AppError {
	var message = ""
	if len(args) > 0 {
		message = tools.FormatJson(args...)
		logo.GetLogger().Info(message)
	}

	return &AppError{
		Code:    code,
		Message: message,
	}
}
