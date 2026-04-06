package authorizationv1

import "errors"

var (
	ErrPermissionDenied error = errors.New("permission denied")
)
