package seed

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/permissionbindings"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/rolebindings"
	"github.com/ooqls/go-auth/v1/roles"
	"github.com/ooqls/go-auth/v1/users"
	"go.uber.org/zap"
)

type Seed struct {
	Users []UserSeed `yaml:"users"`
}

type UserSeed struct {
	Email    string     `yaml:"email"`
	Password string     `yaml:"password"`
	Username string     `yaml:"username"`
	Roles    []RoleSeed `yaml:"roles"`
}

type RoleSeed struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Permissions []string `yaml:"permissions"`
	Hierarchy   int32    `yaml:"hierarchy"`
}

type Service interface {
	Seed(ctx *authorizationv1.Context, seed Seed) error
}

type ServiceImpl struct {
	rolesService              roles.Service
	roleBindingsService       rolebindings.Service
	usersService              users.Service
	permissionsService        permissions.Service
	permissionBindingsService permissionbindings.Service
}

func LoadSeedFromFile(path string) (*Seed, error) {
	seed, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file: %w", err)
	}
	var s Seed
	err = yaml.Unmarshal(seed, &s)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal seed file: %w", err)
	}
	return &s, nil
}

func NewServiceImpl(factory datav1.Factory) Service {
	return &ServiceImpl{
		rolesService:              roles.NewServiceImpl(factory),
		roleBindingsService:       rolebindings.NewServiceImpl(factory),
		usersService:              users.NewServiceImpl(factory),
		permissionsService:        permissions.NewServiceImpl(factory),
		permissionBindingsService: permissionbindings.NewServiceImpl(factory),
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

		createdPerms := make(map[string]struct{})
		for _, role := range user.Roles {
			roleId, err := s.rolesService.CreateRole(ctx, role.Name, role.Description, role.Hierarchy)
			if err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}

			for _, perm := range role.Permissions {

				if _, exists := createdPerms[perm]; !exists {
					ctx.L().Info("creating permission", zap.String("permission", perm))
					err := s.permissionsService.AddPermission(ctx, perm)
					if _, ok := err.(*v1.AlreadyExistsError); !ok && err != nil {
						return fmt.Errorf("failed to add permission: %w", err)
					}
					createdPerms[perm] = struct{}{}
				}

				err = s.permissionBindingsService.AssignPermission(ctx, []uuid.UUID{*roleId}, role.Permissions)
				if err != nil {
					return fmt.Errorf("failed to assign permission to role: %w", err)
				}
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
