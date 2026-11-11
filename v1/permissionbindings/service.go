package permissionbindings

import (
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
)

type Service interface {
	AssignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissions []string) error
	UnassignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissions []string) error
	GetPermissionBindingsForRole(ctx *authorizationv1.Context, roleId uuid.UUID) ([]permissionbindingsv1.Permissionbindingv1, error)
	GetPermissionsBindings(ctx *authorizationv1.Context, page, pageSize int) ([]permissionbindingsv1.Permissionbindingv1, error)
}

type ServiceImpl struct {
	auth authorizationv1.Authorizer
	r    permissionbindingsv1.Reader
	m    bindingMutator
}

func NewServiceImpl(factory datav1.Factory) Service {
	bindings := factory.NewCache("permission_bindings", time.Hour*24)
	permissions := factory.NewCache("permissions", time.Hour*24)
	aggs := factory.NewCache("aggs_reader", time.Hour*24)
	return &ServiceImpl{
		auth: authorizationv1.NewAuthorizerImpl(factory),
		r:    factory.NewPermissionBindingReader(),
		m:    &mutator{transactionRunner: newSQLTransaction(factory, &bindings, &permissions, &aggs)},
	}
}

func (s *ServiceImpl) AssignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissions []string) error {
	return s.m.AssignPermissions(ctx, roleIds, permissions)
}

func (s *ServiceImpl) GetPermissionBindingsForRole(ctx *authorizationv1.Context, roleId uuid.UUID) ([]permissionbindingsv1.Permissionbindingv1, error) {
	if err := s.auth.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.ReadAction, corev1.PermissionsV1.Group, corev1.PermissionsV1.Kind); err != nil {
		return nil, err
	}
	return s.r.GetPermissionBindingsForRole(ctx, roleId)
}

func (s *ServiceImpl) GetPermissionsBindings(ctx *authorizationv1.Context, page, pageSize int) ([]permissionbindingsv1.Permissionbindingv1, error) {
	if err := s.auth.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.ReadAction, corev1.PermissionsV1.Group, corev1.PermissionsV1.Kind); err != nil {
		return nil, err
	}
	return s.r.GetPermissionsBindings(ctx, page, pageSize)
}

func (s *ServiceImpl) UnassignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissions []string) error {
	return s.m.UnassignPermissions(ctx, roleIds, permissions)
}
