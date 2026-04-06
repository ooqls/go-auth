package v1

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrWrongHashAlgo       error = errors.New("password digest was created using the incorrect hashing algorithm")
	ErrNewPasswordRequired error = errors.New("new password is required")
	ErrInvalidEvent        error = errors.New("encountered an invalid event")
)

type M map[string]interface{}

type ErrorCode int

const (
	ErrorCodeNotFound ErrorCode = iota
	ErrorCodeAlreadyExists
	ErrorCodeInternal
	ErrorCodePermissionDenied
)

func ErrNotFound(err error, errorData M) *NotFoundError {
	return &NotFoundError{
		ServiceError: ServiceError{
			error:       err,
			UserMessage: "Not found",
			ErrorCode:   ErrorCodeNotFound,
			ErrorData:   errorData,
		},
	}
}

type NotFoundError struct {
	ServiceError
}

func ErrAlreadyExists(err error, errorData M) *AlreadyExistsError {
	return &AlreadyExistsError{
		ServiceError: ServiceError{
			error:       err,
			UserMessage: "Already exists",
			ErrorCode:   ErrorCodeAlreadyExists,
			ErrorData:   errorData,
		},
	}
}

type AlreadyExistsError struct {
	ServiceError
}

func ErrInternal(err error, errorData M) *InternalError {
	return &InternalError{
		ServiceError: ServiceError{
			error:       err,
			UserMessage: "Internal error",
			ErrorCode:   ErrorCodeInternal,
			ErrorData:   errorData,
		},
	}
}

type InternalError struct {
	ServiceError
}

func ErrPermissionDenied(err error, errorData M) *ForbiddenError {
	return &ForbiddenError{
		ServiceError: ServiceError{
			error:       err,
			UserMessage: "Permission denied",
			ErrorCode:   ErrorCodePermissionDenied,
			ErrorData:   errorData,
		},
	}
}

type ForbiddenError struct {
	ServiceError
}

type ServiceError struct {
	error
	UserMessage string
	ErrorCode   ErrorCode
	ErrorData   M
}

func (e *ServiceError) Error() string {
	return e.error.Error()
}

func (e *ServiceError) Unwrap() error {
	return e.error
}

func GinHandleError(ctx *gin.Context, err error) {
	var notFound *NotFoundError
	var alreadyExists *AlreadyExistsError
	var forbidden *ForbiddenError
	var internal *InternalError

	switch {
	case errors.As(err, &notFound):
		ctx.JSON(404, gin.H{"error": err.Error()})
	case errors.As(err, &alreadyExists):
		ctx.JSON(409, gin.H{"error": err.Error()})
	case errors.As(err, &forbidden):
		ctx.JSON(403, gin.H{"error": err.Error()})
	case errors.As(err, &internal):
		ctx.JSON(500, gin.H{"error": err.Error()})
	default:
		ctx.JSON(500, gin.H{"error": "Internal server error"})
	}
}
