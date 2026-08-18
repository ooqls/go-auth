package permissionbindings

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
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
	w    permissionbindingsv1.Writer
	pr   permissionsv1.Reader
	pw   permissionsv1.Writer
}

func NewServiceImpl(factory datav1.Factory) Service {
	return &ServiceImpl{
		auth: authorizationv1.NewAuthorizerImpl(factory),
		r:    factory.NewPermissionBindingReader(),
		w:    factory.NewPermissionBindingWriter(),
		pr:   factory.NewPermissionReader(),
		pw:   factory.NewPermissionWriter(),
	}
}

func (s *ServiceImpl) AssignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissions []string) error {
	for _, roleId := range roleIds {
		if err := s.auth.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.UpdateAction, corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
			return err
		}

		for _, permission := range permissions {
			if err := s.auth.IsAuthorizedToAssignPermission(ctx, roleId, permission); err != nil {
				return err
			}

			if err := s.w.AssignPermission(ctx, roleId, permission); err != nil {
				return err
			}
		}
	}
	return nil
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
	for _, roleId := range roleIds {
		if err := s.auth.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.UpdateAction, corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
			return err
		}

		for _, permission := range permissions {
			if err := s.auth.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.UnassignAction, corev1.PermissionsV1.Group, corev1.PermissionsV1.Kind); err != nil {
				return err
			}

			if err := s.w.UnassignPermission(ctx, roleId, permission); err != nil {
				return err
			}
		}
	}
	return nil
}
