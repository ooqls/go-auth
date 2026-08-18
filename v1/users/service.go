package users

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/usersv1"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/roles"
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

type User = usersv1.User
type UserId = uuid.UUID

type userCreationOpts struct {
	key   string
	value any
}

var (
	WithKeyOption           = "with-key"
	WithEmailPasswordOption = "with-email-password"
)

func WithKey(key string) userCreationOpts {
	return userCreationOpts{
		key: key,
	}
}

type Service interface {
	CreateUserWithRandomPassword(ctx *authorizationv1.Context, email string, username string) (*User, string, error)
	CreateUserWithPassword(ctx *authorizationv1.Context, email string, username string, key []byte, salt []byte) (*User, error)
	GetUser(ctx *authorizationv1.Context, id UserId) (*User, error)
	GetUsers(ctx *authorizationv1.Context, page int, pageSize int) (*corev1.Result[[]User], error)
	GetUserByUsername(ctx *authorizationv1.Context, username string) (*User, error)
	UpdateUserEmail(ctx *authorizationv1.Context, id UserId, email string) error
	UpdateUserKey(ctx *authorizationv1.Context, id UserId, key []byte) error
	DeleteUser(ctx *authorizationv1.Context, id UserId) error
}

type UserServiceImpl struct {
	l       *zap.Logger
	ua      authorizationv1.Authorizer
	userR   usersv1.Reader
	roleSvc roles.Service
	userW   usersv1.Writer
}

func NewServiceImpl(
	factory datav1.Factory,
) Service {
	userR := factory.NewUserReader()
	userW := factory.NewUserWriter()

	return &UserServiceImpl{
		l:     log.NewLogger("users"),
		ua:    authorizationv1.NewAuthorizerImpl(factory),
		userR: userR,
		userW: userW,
	}
}

func (u *UserServiceImpl) CreateUserWithRandomPassword(ctx *authorizationv1.Context, email string, username string) (*User, string, error) {
	keyBytes, alphanumeric, err := authenticationv1.GenerateKey(authenticationv1.SupportedAlgorithmAESGCM)
	if err != nil {
		u.l.Error("failed to generate key", zap.Error(err))
		return nil, "", v1.ErrInternal(err, v1.M{"username": username})
	}

	salt := authenticationv1.GenerateSalt(username)

	user, err := u.CreateUserWithPassword(ctx, email, username, keyBytes, salt[:])
	if err != nil {
		return nil, "", err
	}

	err = u.userR.ClearCache(ctx)
	if err != nil {
		u.l.Error("failed to clear user cache after creating user with random password", zap.Error(err))
	}

	return user, alphanumeric, nil
}

func (u *UserServiceImpl) CreateUserWithPassword(ctx *authorizationv1.Context, email string, username string, key []byte, salt []byte) (*User, error) {
	if err := u.ua.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.CreateAction, corev1.UsersV1.Group, corev1.UsersV1.Kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"username": username})
	}

	if err := crypto.VerifyGCMAESKey([]byte(key)); err != nil {
		u.l.Error("failed to verify key", zap.Error(err))
		return nil, ErrInvalidKey
	}

	dbUser, err := u.userW.CreateUser(ctx, datagen.CreateUserParams{
		ID:       uuid.New(),
		Username: username,
		Email:    email,
		Key:      []byte(key),
		Salt:     []byte(salt),
	})
	if err != nil {
		u.l.Error("failed to create user", zap.Error(err))
		return nil, v1.ErrInternal(err, v1.M{"username": username})
	}

	err = u.userR.ClearCache(ctx)
	if err != nil {
		u.l.Error("failed to clear user cache after creating user", zap.Error(err))
	}

	user := usersv1.FromDatagenUser(*dbUser)
	return &user, nil
}

func (u *UserServiceImpl) GetUser(ctx *authorizationv1.Context, id uuid.UUID) (*User, error) {
	if err := u.ua.IsAuthorizedToPerformCoreAction(ctx,
		authorizationv1.ReadAction,
		corev1.UsersV1.Group, corev1.UsersV1.Kind,
	); err != nil {
		return nil, err
	}

	user, err := u.userR.GetUser(ctx, id)
	if err != nil {
		u.l.Error("failed to get user", zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (u *UserServiceImpl) GetUsers(ctx *authorizationv1.Context, page int, pageSize int) (*corev1.Result[[]User], error) {
	users, err := u.userR.GetUsers(ctx, int32(page*pageSize), int32(pageSize))
	if err != nil {
		u.l.Error("failed to get users", zap.Error(err))
		return nil, err
	}
	return users, nil
}

// func (u *UserServiceImpl) GetUserAgg(ctx authorization.Context, id records.UserId) (*records.UserAgg, error) {
// 	if err := u.ua.IsAuthorizedToPerformUserAction(&ctx, authorization.ReadAction, users.User{ID: id}); err != nil {
// 		return nil, err
// 	}

// 	userAgg, err := u.userAR.GetUserAgg(&ctx, id)
// 	if err != nil {
// 		u.l.Error("failed to get user agg", zap.String("user_id", id.String()), zap.Error(err))
// 		return nil, err
// 	}

// 	if userAgg == nil {
// 		return nil, nil
// 	}

// 	userAgg.Roles = slices.DeleteFunc(userAgg.Roles, func(role records.RoleAgg) bool {
// 		recordRole := records.Role{
// 			ID:          role.RoleId,
// 			Name:        role.Name,
// 			Hierarchy:   int32(role.Hierarchy),
// 			Description: role.Description,
// 		}
// 		if err := u.ra.IsAuthorizedToPerformRoleAction(ctx, authorization.ReadAction, recordRole); err != nil {
// 			return true
// 		}

// 		return false
// 	})

// 	return userAgg, nil
// }

func (u *UserServiceImpl) GetUserByUsername(ctx *authorizationv1.Context, username string) (*User, error) {

	if err := u.ua.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.ReadAction, corev1.UsersV1.Group, corev1.UsersV1.Kind); err != nil {
		return nil, err
	}

	user, err := u.userR.GetUserByUsername(ctx, username)
	if err != nil {
		u.l.Error("failed to get user by username", zap.Error(err))
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return user, nil
}

func (u *UserServiceImpl) UpdateUserEmail(ctx *authorizationv1.Context, id uuid.UUID, email string) error {
	if err := u.ua.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.UpdateAction, corev1.UsersV1.Group, corev1.UsersV1.Kind); err != nil {
		return err
	}

	curUser, err := u.userR.GetUser(ctx, id)
	if err != nil {
		u.l.Error("failed to get user", zap.Error(err))
		return err
	}

	if curUser == nil {
		return v1.ErrNotFound(err, v1.M{
			"user_id": id,
		})
	}

	if err := u.userW.UpdateUser(ctx, id, datagen.UpdateUserParams{
		Email: pgtype.Text{
			String: email,
			Valid:  true,
		},
	}); err != nil {
		u.l.Error("failed to update user", zap.Error(err))
		return err
	}

	return nil
}

func (u *UserServiceImpl) UpdateUserKey(ctx *authorizationv1.Context, id uuid.UUID, key []byte) error {

	if ctx.User.Id != id {
		if err := u.ua.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.UpdateKeyAction, corev1.UsersV1.Group, corev1.UsersV1.Kind); err != nil {
			return err
		}
	}

	user, err := u.userR.GetUser(ctx, id)
	if err != nil {
		u.l.Error("failed to get user", zap.Error(err))
		return err
	}

	if user == nil {
		return v1.ErrNotFound(err, v1.M{
			"user_id": id,
		})
	}

	err = crypto.VerifyGCMAESKey(key)
	if err != nil {
		u.l.Error("failed to verify key", zap.Error(err))
		return ErrInvalidKey
	}

	curUser, err := u.userR.GetUser(ctx, id)
	if err != nil {
		u.l.Error("failed to get user", zap.Error(err))
		return err
	}

	if curUser == nil {
		return v1.ErrNotFound(err, v1.M{
			"user_id": id,
		})
	}

	curUser.Key = key
	curUser.UpdatedAt = time.Now()

	if err := u.userW.UpdateUser(ctx, id, datagen.UpdateUserParams{
		Key: key,
	}); err != nil {
		u.l.Error("failed to update user key", zap.Error(err))
		return err
	}

	return nil
}

func (u *UserServiceImpl) DeleteUser(ctx *authorizationv1.Context, id uuid.UUID) error {
	if err := u.ua.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.DeleteAction, corev1.UsersV1.Group, corev1.UsersV1.Kind); err != nil {
		return err
	}

	return u.userW.DeleteUser(ctx, id)
}
