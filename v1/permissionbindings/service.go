package permissionbindings

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
	"github.com/ooqls/go-auth/internal/rolesv1"
)

type Service interface {
	AssignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissionIds []uuid.UUID) error
	UnassignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissionIds []uuid.UUID) error
	GetPermissionBindingsForRole(ctx *authorizationv1.Context, roleId uuid.UUID) ([]permissionbindingsv1.Permissionbindingv1, error)
	GetPermissionsBindings(ctx *authorizationv1.Context, page, pageSize int) ([]permissionbindingsv1.Permissionbindingv1, error)
}

type ServiceImpl struct {
	auth authorizationv1.Authorizer
	r    permissionbindingsv1.Reader
	w    permissionbindingsv1.Writer
}

func NewServiceImpl(auth authorizationv1.Authorizer, factory datav1.Factory) *ServiceImpl {
	return &ServiceImpl{
		auth: auth,
		r:    factory.NewPermissionBindingReader(),
		w:    factory.NewPermissionBindingWriter(),
	}
}

func (s *ServiceImpl) AssignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissionIds []uuid.UUID) error {
	for _, roleId := range roleIds {
		role := corev1.FromID(roleId, rolesv1.Metadata)
		if err := s.auth.IsAuthorizedToPerformAction(ctx, authorizationv1.UpdateAction, role); err != nil {
			return err
		}

		for _, permissionId := range permissionIds {
			if err := s.auth.IsAuthorizedToAssignPermission(ctx, roleId, permissionId); err != nil {
				return err
			}

			if err := s.w.AssignPermission(ctx, roleId, permissionId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ServiceImpl) UnassignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissionIds []uuid.UUID) error {
	for _, roleId := range roleIds {
		role := corev1.FromID(roleId, rolesv1.Metadata)
		if err := s.auth.IsAuthorizedToPerformAction(ctx, authorizationv1.UpdateAction, role); err != nil {
			return err
		}

		for _, permissionId := range permissionIds {
			if err := s.auth.IsAuthorizedToUnassignPermission(ctx, roleId); err != nil {
				return err
			}

			if err := s.w.UnassignPermission(ctx, roleId, permissionId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ServiceImpl) GetPermissionBindingsForRole(ctx *authorizationv1.Context, roleId uuid.UUID) ([]permissionbindingsv1.Permissionbindingv1, error) {
	if err := s.auth.IsAuthorizedToReadRolePermissions(ctx, roleId); err != nil {
		return nil, err
	}

	return s.r.GetPermissionBindingsForRole(ctx, roleId)
}

func (s *ServiceImpl) GetPermissionsBindings(ctx *authorizationv1.Context, page, pageSize int) ([]permissionbindingsv1.Permissionbindingv1, error) {
	if err := s.auth.IsAuthorizedToPerformGlobalAction(ctx, authorizationv1.ReadAction, permissionbindingsv1.Metadata); err != nil {
		return nil, authorizationv1.ErrPermissionDenied
	}

	return s.r.GetPermissionsBindings(ctx, page, pageSize)
}
