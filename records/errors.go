package records

import "errors"

var (
	ErrWrongHashAlgo       error = errors.New("password digest was created using the incorrect hashing algorithm")
	ErrNewPasswordRequired error = errors.New("new password is required")
	ErrInvalidEvent        error = errors.New("encountered an invalid event")
)
