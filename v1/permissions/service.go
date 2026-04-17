package permissions

import (
	"slices"
	"strings"

	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/pkg/errors"
)

var _ Service = &ServiceImpl{}

type Service interface {
	AddPermission(ctx *authorizationv1.Context, name, group, kind string, actions []string) error
	DeletePermission(ctx *authorizationv1.Context, name, group, kind string) error
	GetPermissions(ctx *authorizationv1.Context, page int, pageSize int) ([]permissionsv1.Permission, error)
	GetPermission(ctx *authorizationv1.Context, name, group, kind string) (*permissionsv1.Permission, error)
}

type ServiceImpl struct {
	pr permissionsv1.Reader
	pw permissionsv1.Writer
	ra authorizationv1.Authorizer
}

func NewServiceImpl(factory datav1.Factory) *ServiceImpl {
	return &ServiceImpl{
		pr: factory.NewPermissionReader(),
		pw: factory.NewPermissionWriter(),
		ra: authorizationv1.NewAuthorizerImpl(factory),
	}
}

func (s *ServiceImpl) AddPermission(ctx *authorizationv1.Context, name, group, kind string, actions []string) error {
	perm := permissionsv1.NewPermission(group, kind, name, strings.Join(actions, ","))
	if err := s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.CreateAction, perm.Object); err != nil {
		return v1.ErrPermissionDenied(err, v1.M{"name": name, "group": group, "kind": kind})
	}

	return s.pw.CreatePermission(name, group, kind, actions)
}

func (s *ServiceImpl) DeletePermission(ctx *authorizationv1.Context, name, group, kind string) error {
	target := permissionsv1.NewPermission(group, kind, name, "")
	if err := s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.DeleteAction, target.Object); err != nil {
		return v1.ErrPermissionDenied(err, v1.M{"name": name, "group": group, "kind": kind})
	}

	return s.pw.DeletePermission(name, group, kind)
}

func (s *ServiceImpl) GetPermissions(ctx *authorizationv1.Context, page int, pageSize int) ([]permissionsv1.Permission, error) {
	permissions, err := s.pr.GetPermissions(ctx, page, pageSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get permissions")
	}

	permissions = slices.DeleteFunc(permissions, func(p permissionsv1.Permission) bool {
		target := corev1.Object{
			Metadata: corev1.PermissionsV1,
			Name:     p.Name,
		}
		return s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.ReadAction, target) != nil
	})

	return permissions, nil
}

func (s *ServiceImpl) GetPermission(ctx *authorizationv1.Context, name, group, kind string) (*permissionsv1.Permission, error) {
	permission, err := s.pr.GetPermission(ctx, name, group, kind)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get permission")
	}

	if permission == nil {
		return nil, v1.ErrNotFound(errors.New("permission not found"), v1.M{"name": name, "group": group, "kind": kind})
	}

	target := corev1.Object{
		Metadata: corev1.PermissionsV1,
		Name:     permission.Name,
	}
	if err := s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.ReadAction, target); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"name": name, "group": group, "kind": kind})
	}

	return permission, nil
}
