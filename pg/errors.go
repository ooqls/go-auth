package pg

import "errors"

var (
	ErrTooManyRows error = errors.New("too many rows returned, only one expected")
)
