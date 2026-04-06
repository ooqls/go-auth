package rolebindings

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1"
	"go.uber.org/zap"
)

type Service interface {
	GetRoleBindingsForUser(ctx *authorizationv1.Context, userId uuid.UUID) ([]rolebindingsv1.Rolebinding, error)
	AssignRoleToUser(ctx *authorizationv1.Context, userId uuid.UUID, roleId uuid.UUID) error
	UnassignRoleFromUser(ctx *authorizationv1.Context, userId uuid.UUID, roleId uuid.UUID) error
}

type ServiceImpl struct {
	auth authorizationv1.Authorizer
	rr   rolebindingsv1.Reader
	rw   rolebindingsv1.Writer
}

func NewServiceImpl(auth authorizationv1.Authorizer, rr rolebindingsv1.Reader, rw rolebindingsv1.Writer) *ServiceImpl {
	return &ServiceImpl{auth: auth, rr: rr, rw: rw}
}

func (s *ServiceImpl) GetRoleBindingsForUser(ctx *authorizationv1.Context, userId uuid.UUID) ([]rolebindingsv1.Rolebinding, error) {
	if err := s.auth.IsAuthorizedToReadUserRoles(ctx, userId); err != nil {
		return nil, err
	}

	return s.rr.GetRoleBindingsForUser(ctx, userId)
}

func (s *ServiceImpl) AssignRoleToUser(ctx *authorizationv1.Context, userId uuid.UUID, roleId uuid.UUID) error {
	return s.rw.AddRoleToUser(ctx, userId, roleId)
}

func (s *ServiceImpl) UnassignRoleFromUser(ctx *authorizationv1.Context, userId uuid.UUID, roleId uuid.UUID) error {
	if err := s.rr.ClearCacheForUser(ctx, userId); err != nil {
		ctx.L().Warn("failed to clear cache for reader", zap.Error(err))
	}

	return s.rw.RemoveRoleFromUser(ctx, userId, roleId)
}
