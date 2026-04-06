package aggs

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/aggsv1"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/pkg/errors"
)

var _ Service = &ServiceImpl{}

type Service interface {
	GetRoleAgg(ctx *authorizationv1.Context, roleID uuid.UUID) (*aggsv1.RoleAgg, error)
	GetUserAgg(ctx *authorizationv1.Context, userID uuid.UUID) (*aggsv1.UserAgg, error)
	GetUserHierarchy(ctx *authorizationv1.Context, userID uuid.UUID) (int32, error)
	CompareUserHierarchy(ctx *authorizationv1.Context, userA, userB uuid.UUID) (int32, error)
}

type ServiceImpl struct {
	ar aggsv1.Reader
	ra authorizationv1.Authorizer
}

func NewServiceImpl(factory datav1.Factory) *ServiceImpl {
	return &ServiceImpl{
		ar: factory.NewAggReader(),
		ra: authorizationv1.NewAuthorizerImpl(factory),
	}
}

func (s *ServiceImpl) GetRoleAgg(ctx *authorizationv1.Context, roleID uuid.UUID) (*aggsv1.RoleAgg, error) {
	if err := s.ra.IsAuthorizedToReadRolePermissions(ctx, roleID); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"role_id": roleID})
	}

	agg, err := s.ar.GetRoleAgg(ctx, roleID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get role aggregate")
	}

	if agg == nil {
		return nil, v1.ErrNotFound(errors.New("role not found"), v1.M{"role_id": roleID})
	}

	return agg, nil
}

func (s *ServiceImpl) GetUserAgg(ctx *authorizationv1.Context, userID uuid.UUID) (*aggsv1.UserAgg, error) {
	if err := s.ra.IsAuthorizedToReadUserRoles(ctx, userID); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"user_id": userID})
	}

	agg, err := s.ar.GetUserAgg(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user aggregate")
	}

	if agg == nil {
		return nil, v1.ErrNotFound(errors.New("user not found"), v1.M{"user_id": userID})
	}

	return agg, nil
}

func (s *ServiceImpl) GetUserHierarchy(ctx *authorizationv1.Context, userID uuid.UUID) (int32, error) {
	if err := s.ra.IsAuthorizedToReadUserRoles(ctx, userID); err != nil {
		return 0, v1.ErrPermissionDenied(err, v1.M{"user_id": userID})
	}

	hierarchy, err := s.ar.GetUserHierarchy(ctx, userID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get user hierarchy")
	}

	return hierarchy, nil
}

func (s *ServiceImpl) CompareUserHierarchy(ctx *authorizationv1.Context, userA, userB uuid.UUID) (int32, error) {
	if err := s.ra.IsAuthorizedToReadUserRoles(ctx, userA); err != nil {
		return 0, v1.ErrPermissionDenied(err, v1.M{"user_id": userA})
	}

	if err := s.ra.IsAuthorizedToReadUserRoles(ctx, userB); err != nil {
		return 0, v1.ErrPermissionDenied(err, v1.M{"user_id": userB})
	}

	result, err := s.ar.CompareUserHierarchy(ctx, userA, userB)
	if err != nil {
		return 0, errors.Wrap(err, "failed to compare user hierarchies")
	}

	return result, nil
}
