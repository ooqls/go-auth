package permissions

import (
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/pkg/errors"
)

var _ Service = &ServiceImpl{}

type Service interface {
	AddPermission(ctx *authorizationv1.Context, permission string) error
	DeletePermission(ctx *authorizationv1.Context, permission string) error
	GetPermissions(ctx *authorizationv1.Context, page int, pageSize int) ([]permissionsv1.Permission, error)
	GetPermission(ctx *authorizationv1.Context, permission string) (*permissionsv1.Permission, error)
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

func (s *ServiceImpl) AddPermission(ctx *authorizationv1.Context, permission string) error {
	if err := s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.CreateAction, corev1.ToTargetString(permissionsv1.Metadata)); err != nil {
		return v1.ErrPermissionDenied(err, v1.M{"permission": permission})
	}

	retrPerm, err := s.pr.GetPermission(ctx, permission)
	if err != nil {
		return errors.Wrap(err, "failed to check for existing permission")
	}

	if retrPerm != nil {
		return v1.ErrAlreadyExists(errors.New("permission already exists"), v1.M{"permission": permission})
	}

	_, err = s.pw.CreatePermission(ctx, permission)
	if err != nil {
		return v1.ErrInternal(err, v1.M{"permission": permission})
	}
	return nil
}

func (s *ServiceImpl) DeletePermission(ctx *authorizationv1.Context, permission string) error {
	if err := s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.DeleteAction, corev1.ToTargetString(permissionsv1.Metadata)); err != nil {
		return v1.ErrPermissionDenied(err, v1.M{"permission": permission})
	}

	return s.pw.DeletePermission(ctx, permission)
}

func (s *ServiceImpl) GetPermissions(ctx *authorizationv1.Context, page int, pageSize int) ([]permissionsv1.Permission, error) {
	if err := s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.ReadAction, corev1.ToTargetString(permissionsv1.Metadata)); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{})
	}

	permissions, err := s.pr.GetPermissions(ctx, page, pageSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get permissions")
	}

	return permissions, nil
}

func (s *ServiceImpl) GetPermission(ctx *authorizationv1.Context, permission string) (*permissionsv1.Permission, error) {
	if err := s.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.ReadAction, corev1.ToTargetString(permissionsv1.Metadata)); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"permission": permission})
	}

	retrPerm, err := s.pr.GetPermission(ctx, permission)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get permission")
	}

	if retrPerm == nil {
		return nil, v1.ErrNotFound(errors.New("permission not found"), v1.M{})
	}

	return retrPerm, nil
}
