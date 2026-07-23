package apperror

import (
	"net/http"

	"github.com/pkg/errors"
)

var (
	ErrNotFound               = errors.New("not found")
	ErrUnprocessableEntity    = errors.New("unprocessable entity")
	ErrForbiddenAccess        = errors.New("forbidden access")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrConflict               = errors.New("conflict")
	ErrTimeout                = errors.New("timeout")
	ErrBadRequest             = errors.New("bad request")
	ErrGateway                = errors.New("gateway")
	ErrBodyMustNonPointerType = errors.New("body must be a non-pointer type")
)

type AppError struct {
	Code    int
	Err     error
	Message string
}

func (h AppError) Error() string {
	return h.Err.Error()
}

func BadRequest(err error) error {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: "bad_request",
		Err:     err,
	}
}

func InternalServerError(err error) error {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "internal_server_error",
		Err:     err,
	}
}

func Unauthorized(err error) error {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: "unauthorized",
		Err:     err,
	}
}

func Forbidden(err error) error {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: "forbidden",
		Err:     err,
	}
}

func NotFound(err error) error {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: "not_found",
		Err:     err,
	}
}

func Conflict(err error) error {
	return &AppError{
		Code:    http.StatusConflict,
		Message: "error conflict! must rollback",
		Err:     err,
	}
}

func GatewayTimeout(err error) error {
	return &AppError{
		Code:    http.StatusGatewayTimeout,
		Message: "gateway_timeout",
		Err:     err,
	}
}

func UnprocessableEntity(err error) error {
	return &AppError{
		Code:    http.StatusUnprocessableEntity,
		Message: "unprocessable_entity",
		Err:     err,
	}
}
