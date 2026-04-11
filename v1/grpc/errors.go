package grpc

import (
	"errors"

	v1 "github.com/ooqls/go-auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleError maps service-layer errors to gRPC status errors.
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	var notFound *v1.NotFoundError
	var alreadyExists *v1.AlreadyExistsError
	var forbidden *v1.ForbiddenError
	var internal *v1.InternalError

	switch {
	case errors.As(err, &notFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.As(err, &alreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.As(err, &forbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.As(err, &internal):
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
