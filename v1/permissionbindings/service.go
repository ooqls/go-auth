package permissionbindings

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
)

type Service interface {
	AssignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissionIds []uuid.UUID) error
	UnassignPermission(ctx *authorizationv1.Context, roleIds []uuid.UUID, permissionIds []uuid.UUID) error
}

type ServiceImpl struct {
	r permissionbindingsv1.Reader
	w permissionbindingsv1.Writer
}

func NewServiceImpl(factory datav1.Factory) *ServiceImpl {
	return &ServiceImpl{
		r: factory.NewPermissionBindingReader(),
		w: factory.NewPermissionBindingWriter(),
	}
}
