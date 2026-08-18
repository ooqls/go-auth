package rolebindings

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
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

func NewServiceImpl(data datav1.Factory) *ServiceImpl {
	return &ServiceImpl{
		auth: authorizationv1.NewAuthorizerImpl(data),
		rr:   data.NewRoleBindingsReader(),
		rw:   data.NewRoleBindingsWriter(),
	}
}

func (s *ServiceImpl) GetRoleBindingsForUser(ctx *authorizationv1.Context, userId uuid.UUID) ([]rolebindingsv1.Rolebinding, error) {
	if err := s.auth.IsAuthorizedToReadUserRoles(ctx, userId); err != nil {
		return nil, err
	}

	return s.rr.GetRoleBindingsForUser(ctx, userId)
}

func (s *ServiceImpl) AssignRoleToUser(ctx *authorizationv1.Context, userId uuid.UUID, roleId uuid.UUID) error {
	if err := s.auth.IsAuthorizedToModifyUser(ctx, userId); err != nil {
		return err
	}

	return s.rw.AddRoleToUser(ctx, userId, roleId)
}

func (s *ServiceImpl) UnassignRoleFromUser(ctx *authorizationv1.Context, userId uuid.UUID, roleId uuid.UUID) error {
	if err := s.auth.IsAuthorizedToModifyUser(ctx, userId); err != nil {
		return err
	}

	if err := s.rr.ClearCacheForUser(ctx, userId); err != nil {
		ctx.L().Warn("failed to clear cache for reader", zap.Error(err))
	}
	return s.rw.RemoveRoleFromUser(ctx, userId, roleId)
}
