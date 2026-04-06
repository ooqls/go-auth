package resources

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	"go.uber.org/zap"
)

type Service interface {
	CreateResource(ctx authorizationv1.Context, group, kind, name string) (Resourcev1, error)
	GetResource(ctx authorizationv1.Context, id resourceIdentifier) (Resourcev1, error)
	UpdateResourceName(ctx authorizationv1.Context, id uuid.UUID, name string, description *string) (Resourcev1, error)
	DeleteResource(ctx authorizationv1.Context, id resourceIdentifier) error
	ListResources(ctx authorizationv1.Context, group, kind string, page, pageSize int32) ([]Resourcev1, error)
}

type ServiceImpl struct {
	rr resourcesv1.Reader
	rw resourcesv1.Writer
	ra authorizationv1.Authorizer
}

func NewServiceImpl(factory datav1.Factory) *ServiceImpl {
	return &ServiceImpl{rw: factory.NewResourceWriter(), rr: factory.NewResourceReader(), ra: authorizationv1.NewAuthorizerImpl(factory)}
}

func (s *ServiceImpl) CreateResource(ctx authorizationv1.Context, group, kind, name, description string) (*Resourcev1, error) {
	res, err := s.rw.CreateResource(ctx, group, kind, name, description)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *ServiceImpl) GetResource(ctx authorizationv1.Context, group, kind, name string) (*Resourcev1, error) {
	res, err := s.rr.GetResource(ctx, name, corev1.Metadata{Group: group, Kind: kind})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *ServiceImpl) UpdateResourceName(ctx authorizationv1.Context, id uuid.UUID, name string, description *string) (*Resourcev1, error) {
	if err := s.rr.ClearCache(ctx); err != nil {
		ctx.L().Warn("failed to clear reader cache", zap.Error(err))
	}

	res, err := s.rw.UpdateResource(ctx, id, &name, description)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *ServiceImpl) DeleteResource(ctx authorizationv1.Context, id uuid.UUID) error {
	if err := s.rr.ClearCache(ctx); err != nil {
		ctx.L().Warn("failed to clear reader cache", zap.Error(err))
	}

	return s.rw.DeleteResourceById(ctx, id)
}

func (s *ServiceImpl) DeleteResourceByName(ctx authorizationv1.Context, group, kind, name string) error {
	if err := s.rr.ClearCache(ctx); err != nil {
		ctx.L().Warn("failed to clear reader cache", zap.Error(err))
	}
	return s.rw.DeleteResource(ctx, group, kind, name)
}

func (s *ServiceImpl) ListResources(ctx authorizationv1.Context, group, kind string, page, pageSize int32) ([]Resourcev1, error) {
	res, err := s.rr.GetResources(ctx, corev1.Metadata{Group: group, Kind: kind}, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}

	return res, nil
}
