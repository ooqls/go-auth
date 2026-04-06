package seed

import (
	"fmt"

	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/v1/rolebindings"
	"github.com/ooqls/go-auth/v1/roles"
	"github.com/ooqls/go-auth/v1/users"
	"go.uber.org/zap"
)

type Seed struct {
	Users []UserSeed
}

type UserSeed struct {
	Email    string
	Password string
	Username string
	Roles    []RoleSeed
}

type RoleSeed struct {
	Name        string
	Description string
	Permissions []PermissionSeed
	Hierarchy   int32
}

type PermissionSeed struct {
	Name        string
	Description string
	Group       string
	Kind        string
	Actions     []string
}

type Service interface {
	Seed(ctx *authorizationv1.Context, seed Seed) error
}

type ServiceImpl struct {
	rolesService        roles.Service
	roleBindingsService rolebindings.Service
	usersService        users.Service
}

func NewServiceImpl(
	rolesService roles.Service,
	roleBindingsService rolebindings.Service,
	usersService users.Service) Service {
	return &ServiceImpl{
		rolesService:        rolesService,
		roleBindingsService: roleBindingsService,
		usersService:        usersService,
	}
}

func (s *ServiceImpl) Seed(ctx *authorizationv1.Context, seed Seed) error {
	for _, user := range seed.Users {
		salt := authenticationv1.GenerateSalt(user.Username)
		key, err := crypto.DeriveAESGCMKey(user.Password, [16]byte(salt))
		if err != nil {
			return fmt.Errorf("failed to derive AES GCM key: %w", err)
		}
		userid, err := s.usersService.CreateUserWithPassword(ctx, user.Email, user.Username, key, salt[:])
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		for _, role := range user.Roles {
			roleId, err := s.rolesService.CreateRole(ctx, roles.CreateRoleParams{
				Name:        role.Name,
				Description: role.Description,
				Hierarchy:   role.Hierarchy,
			})
			if err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}

			ctx.L().Info("assigning role to user", zap.String("role", roleId.String()), zap.String("user", userid.Id.String()))
			err = s.roleBindingsService.AssignRoleToUser(ctx, userid.Id, *roleId)
			if err != nil {
				return fmt.Errorf("failed to assign role to user: %w", err)
			}
		}
	}
	return nil
}
