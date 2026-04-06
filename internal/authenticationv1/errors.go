package authenticationv1

import "errors"

var (
	ErrChallengeExpired    error = errors.New("challenge expired")
	ErrChallengeFailed     error = errors.New("challenge failed")
	ErrInvalidRegistration error = errors.New("invalid registration")
	ErrInternal            error = errors.New("internal error")
)

var (
	ErrTokenExpired        error = errors.New("token expired")
	ErrInvalidToken        error = errors.New("token invalid")
	ErrInvalidPassword     error = errors.New("invalid password")
	ErrUserNotFound        error = errors.New("user not found")
	ErrUserExists          error = errors.New("user already exists")
	ErrRegistrationExpired error = errors.New("registration expired")
)
