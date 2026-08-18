package resources

import (
	"errors"
	"fmt"

	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	v1 "github.com/ooqls/go-auth/v1"
	"go.uber.org/zap"
)

var _ Service = &ServiceImpl{}

type Service interface {
	CreateResource(ctx *authorizationv1.Context, group, kind, name string) (*Resourcev1, error)
	GetResource(ctx *authorizationv1.Context, group, kind, name string) (*Resourcev1, error)
	GetResources(ctx *authorizationv1.Context, group, kind *string, page, pageSize int32) (*corev1.Result[[]Resourcev1], error)
	SearchResources(ctx *authorizationv1.Context, group, kind, name, query *string) (*corev1.Result[[]Resourcev1], error)
	UpdateResourceName(ctx *authorizationv1.Context, group, kind, name string, newName string) (*Resourcev1, error)
	DeleteResource(ctx *authorizationv1.Context, group, kind, name string) error
}

const defaultSearchPageSize = 100

type ServiceImpl struct {
	rr resourcesv1.Reader
	rw resourcesv1.Writer
	ra authorizationv1.Authorizer
}

func NewServiceImpl(factory datav1.Factory) Service {
	return &ServiceImpl{
		rw: factory.NewResourceWriter(),
		rr: factory.NewResourceReader(),
		ra: authorizationv1.NewAuthorizerImpl(factory),
	}
}

func (s *ServiceImpl) CreateResource(ctx *authorizationv1.Context, group, kind, name string) (*Resourcev1, error) {
	if err := s.ra.IsAuthorizedToPerformResourceAction(ctx, authorizationv1.CreateAction, group, kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"group": group, "kind": kind})
	}

	res, err := s.rw.CreateResource(ctx, group, kind, name)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"group": group, "kind": kind, "name": name})
	}

	if err := s.rr.ClearCache(ctx); err != nil {
		ctx.L().Warn("failed to clear cache", zap.Error(err))
	}

	return res, nil
}

func (s *ServiceImpl) GetResource(ctx *authorizationv1.Context, group, kind, name string) (*Resourcev1, error) {
	if err := s.ra.IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, group, kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"group": group, "kind": kind, "name": name})
	}

	res, err := s.rr.GetResource(ctx, group, kind, name)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *ServiceImpl) UpdateResourceName(ctx *authorizationv1.Context, group, kind, name, newName string) (*Resourcev1, error) {
	if err := s.ra.IsAuthorizedToPerformResourceAction(ctx, authorizationv1.UpdateAction, group, kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{
			"group": group,
			"kind":  kind,
			"name":  name,
		})
	}

	if err := s.rr.ClearCache(ctx); err != nil {
		ctx.L().Warn("failed to clear reader cache", zap.Error(err))
	}

	res, err := s.rr.GetResource(ctx, group, kind, name)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, v1.ErrPermissionDenied(fmt.Errorf("resource not found"), v1.M{
			"group": group,
			"kind":  kind,
			"name":  name,
		})
	}

	res, err = s.rw.UpdateResource(ctx, group, kind, name, newName)
	if err != nil {
		if errors.Is(err, resourcesv1.ErrNotFound) {
			return nil, v1.ErrNotFound(fmt.Errorf("resource not found"), v1.M{
				"group": group,
				"kind":  kind,
				"name":  name,
			})
		}

		return nil, v1.ErrInternal(err, v1.M{
			"group": group,
			"kind":  kind,
			"name":  name,
		})
	}

	return res, nil
}

func (s *ServiceImpl) DeleteResource(ctx *authorizationv1.Context, group, kind, name string) error {
	if err := s.ra.IsAuthorizedToPerformResourceAction(ctx, authorizationv1.DeleteAction, group, kind); err != nil {
		return v1.ErrPermissionDenied(err, v1.M{"group": group, "kind": kind, "name": name})
	}

	if err := s.rr.ClearCache(ctx); err != nil {
		ctx.L().Warn("failed to clear reader cache", zap.Error(err))
	}

	err := s.rw.DeleteResource(ctx, group, kind, name)
	if err != nil {
		return v1.ErrInternal(err, v1.M{"group": group, "kind": kind, "name": name})
	}

	return nil
}

func (s *ServiceImpl) DeleteResourceByName(ctx *authorizationv1.Context, group, kind, name string) error {
	if err := s.rr.ClearCache(ctx); err != nil {
		ctx.L().Warn("failed to clear reader cache", zap.Error(err))
	}
	return s.rw.DeleteResource(ctx, group, kind, name)
}

func (s *ServiceImpl) SearchResources(ctx *authorizationv1.Context, group, kind, name, query *string) (*corev1.Result[[]Resourcev1], error) {
	targetGroup := "*"
	targetKind := "*"

	if group != nil {
		targetGroup = *group
	}

	if kind != nil {
		targetKind = *kind
	}

	if err := s.ra.IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, targetGroup, targetKind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"group": targetGroup, "kind": targetKind})
	}

	return s.rr.SearchResources(ctx, group, kind, name, query, defaultSearchPageSize, 0)
}

func (s *ServiceImpl) GetResources(ctx *authorizationv1.Context, group, kind *string, page, pageSize int32) (*corev1.Result[[]Resourcev1], error) {
	targetGroup := "*"
	targetKind := "*"

	if group != nil {
		targetGroup = *group
	}

	if kind != nil {
		targetKind = *kind
	}

	target := ToTargetString(targetGroup, targetKind)
	ctx.L().Info("attempting to read reosurces", zap.String("target", target))
	if err := s.ra.IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, targetGroup, targetKind); err != nil {
		return nil, err
	}
	offset := max(0, (page-1)*pageSize)
	switch {
	case group != nil && kind != nil:
		return s.rr.GetResourcesByGroupAndKind(ctx, *group, *kind, pageSize, offset)
	case group != nil:
		return s.rr.GetResourcesByGroup(ctx, *group, pageSize, offset)
	default:
		return s.rr.GetResources(ctx, pageSize, offset)
	}
}
