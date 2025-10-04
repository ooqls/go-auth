package users

import (
	"errors"

	"github.com/ooqls/go-auth/domain/v1/authorization"
	"github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/users"
	"github.com/ooqls/go-crypto/crypto"
	"github.com/ooqls/go-log"
	"go.uber.org/zap"
)

var (
	ErrInvalidEmail      error = errors.New("invalid email")
	ErrInvalidKey        error = errors.New("invalid key")
	ErrInvalidSalt       error = errors.New("invalid salt")
	ErrInvalidUsername   error = errors.New("invalid username")
	ErrUserAlreadyExists error = errors.New("user already exists")
	ErrInternal          error = errors.New("internal error")
)

type UserService interface {
	CreateUser(ctx authorization.Context, email, key, salt, username string) (*users.User, error)
	GetUser(ctx authorization.Context, id records.UserId) (*users.User, error)
	GetUserByUsername(ctx authorization.Context, username string) (*users.User, error)
	UpdateUser(ctx authorization.Context, id records.UserId, email, key, username string) error
	DeleteUser(ctx authorization.Context, id records.UserId) error
}

type UserServiceImpl struct {
	l     *zap.Logger
	ua    authorization.UserAuthorizer
	userR users.Reader
	userW users.Writer
}

func NewUserServiceImpl(ua authorization.UserAuthorizer, userR users.Reader, userW users.Writer) UserService {
	return &UserServiceImpl{
		l:     log.NewLogger("users"),
		ua:    ua,
		userR: userR,
		userW: userW,
	}
}

func (u *UserServiceImpl) CreateUser(ctx authorization.Context, email, key, salt, username string) (*users.User, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	if salt == "" {
		return nil, ErrInvalidSalt
	}

	if username == "" {
		return nil, ErrInvalidUsername
	}

	err := crypto.VerifyGCMAESKey([]byte(key))
	if err != nil {
		u.l.Error("failed to verify key", zap.Error(err))
		return nil, ErrInvalidKey
	}

	existingUser, err := u.userR.GetUserByUsername(ctx, username)
	if err != nil {
		u.l.Error("failed to get user by username", zap.Error(err))
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	user := users.User{
		Email:    email,
		Key:      []byte(key),
		Username: username,
		Salt:     []byte(salt),
		ID:       records.NewUserID(),
	}

	// err = u.ua.IsAuthorizedToPerformUserAction(&ctx, authorization.CreateAction, user)
	// if err != nil {
	// 	return err
	// }

	err = u.userW.CreateUser(ctx, user)
	if err != nil {
		u.l.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func (u *UserServiceImpl) GetUser(ctx authorization.Context, id records.UserId) (*users.User, error) {
	if err := u.ua.IsAuthorizedToPerformUserAction(&ctx, authorization.ReadAction, users.User{ID: id}); err != nil {
		return nil, err
	}

	return u.userR.GetUser(ctx, id)
}

func (u *UserServiceImpl) GetUserByUsername(ctx authorization.Context, username string) (*users.User, error) {
	user, err := u.userR.GetUserByUsername(ctx, username)
	if err != nil {
		u.l.Error("failed to get user by username", zap.Error(err))
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	if err := u.ua.IsAuthorizedToPerformUserAction(&ctx, authorization.ReadAction, *user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserServiceImpl) UpdateUser(ctx authorization.Context, targetUserid records.UserId, email, key, username string) error {
	if err := u.ua.IsAuthorizedToPerformUserAction(&ctx, authorization.UpdateAction, users.User{ID: targetUserid}); err != nil {
		return err
	}

	return u.userW.UpdateUser(ctx, users.User{
		ID:       targetUserid,
		Email:    email,
		Key:      []byte(key),
		Username: username,
	})
}

func (u *UserServiceImpl) DeleteUser(ctx authorization.Context, id records.UserId) error {
	if err := u.ua.IsAuthorizedToPerformUserAction(&ctx, authorization.DeleteAction, users.User{ID: id}); err != nil {
		return err
	}

	return u.userW.DeleteUser(ctx, id)
}
