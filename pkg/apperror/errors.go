package apperror

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(httpStatus int, code, message string) *AppError {
	return &AppError{HTTPStatus: httpStatus, Code: code, Message: message}
}

var (
	ErrNotFound     = New(http.StatusNotFound, "NOT_FOUND", "resource not found")
	ErrUnauthorized = New(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
	ErrForbidden    = New(http.StatusForbidden, "FORBIDDEN", "forbidden")
	ErrConflict     = New(http.StatusConflict, "CONFLICT", "resource already exists")
	ErrBadRequest   = New(http.StatusBadRequest, "BAD_REQUEST", "bad request")
	ErrInternal     = New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	ErrTokenExpired = New(http.StatusUnauthorized, "TOKEN_EXPIRED", "token has expired")
	ErrTokenInvalid = New(http.StatusUnauthorized, "TOKEN_INVALID", "token is invalid")
)

func Is(err error, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == target.Code
	}
	return false
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
