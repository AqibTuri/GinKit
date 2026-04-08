// Package apperrors defines typed errors services return; pkg/response maps them to JSON + HTTP status.
package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a typed API error with a stable code for clients and an HTTP status for the transport layer.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status, Err: err}
}

// Sentinel errors (compare with errors.Is in services).
var (
	ErrNotFound           = New("NOT_FOUND", "resource not found", http.StatusNotFound)
	ErrConflict           = New("CONFLICT", "resource already exists", http.StatusConflict)
	ErrUnauthorized       = New("UNAUTHORIZED", "unauthorized", http.StatusUnauthorized)
	ErrForbidden          = New("FORBIDDEN", "forbidden", http.StatusForbidden)
	ErrValidation         = New("VALIDATION_ERROR", "validation failed", http.StatusBadRequest)
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "invalid email or password", http.StatusUnauthorized)
)

func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
