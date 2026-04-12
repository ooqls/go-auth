package resourcesgrpc

import (
	"context"

	"github.com/google/uuid"
	resourcespb "github.com/ooqls/go-auth/gen/grpc/resources/v1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ resourcespb.ResourcesServiceServer = (*Server)(nil)

type Server struct {
	resourcespb.UnimplementedResourcesServiceServer
	resourceReader resourcesv1.Reader
	resourceWriter resourcesv1.Writer
	l              *zap.Logger
}

func NewServer(factory datav1.Factory, l *zap.Logger) *Server {
	return &Server{
		resourceReader: factory.NewResourceReader(),
		resourceWriter: factory.NewResourceWriter(),
		l:              l,
	}
}

func (s *Server) GetResource(ctx context.Context, req *resourcespb.GetResourceRequest) (*resourcespb.GetResourceResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	res, err := s.resourceReader.GetResource(ctx, req.GetName(), corev1.Metadata{
		Group: req.GetGroup(),
		Kind:  req.GetKind(),
	})
	if err != nil {
		s.l.Error("failed to get resource", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get resource")
	}

	if res == nil {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	return &resourcespb.GetResourceResponse{
		Resource: grpcutil.ResourceToProto(res),
	}, nil
}

func (s *Server) DeleteResource(ctx context.Context, req *resourcespb.DeleteResourceRequest) (*resourcespb.DeleteResourceResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	err := s.resourceWriter.DeleteResource(ctx, req.GetGroup(), req.GetKind(), req.GetName())
	if err != nil {
		s.l.Error("failed to delete resource", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete resource")
	}

	return &resourcespb.DeleteResourceResponse{}, nil
}

func (s *Server) ListResources(ctx context.Context, req *resourcespb.ListResourcesRequest) (*resourcespb.ListResourcesResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	group := ""
	kind := ""
	page := int32(0)
	pageSize := int32(100)

	if req.Group != nil {
		group = req.GetGroup()
	}
	if req.Kind != nil {
		kind = req.GetKind()
	}
	if req.Page != nil {
		page = req.GetPage()
	}
	if req.PageSize != nil {
		pageSize = req.GetPageSize()
	}

	res, err := s.resourceReader.GetResources(ctx, corev1.Metadata{
		Group: group,
		Kind:  kind,
	}, pageSize, page*pageSize)
	if err != nil {
		s.l.Error("failed to get resources", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get resources")
	}

	return &resourcespb.ListResourcesResponse{
		Resources: grpcutil.ResourcesToProto(res),
	}, nil
}

func (s *Server) CreateResource(ctx context.Context, req *resourcespb.CreateResourceRequest) (*resourcespb.CreateResourceResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	res, err := s.resourceWriter.CreateResource(ctx, req.GetGroup(), req.GetKind(), req.GetName(), req.GetDescription())
	if err != nil {
		s.l.Error("failed to create resource", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create resource")
	}

	return &resourcespb.CreateResourceResponse{
		Resource: grpcutil.ResourceToProto(res),
	}, nil
}

func (s *Server) UpdateResource(ctx context.Context, req *resourcespb.UpdateResourceRequest) (*resourcespb.UpdateResourceResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid resource id")
	}

	name := req.GetName()
	var desc *string
	if req.Description != nil {
		d := req.GetDescription()
		desc = &d
	}

	res, err := s.resourceWriter.UpdateResource(ctx, id, &name, desc)
	if err != nil {
		s.l.Error("failed to update resource", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update resource")
	}

	return &resourcespb.UpdateResourceResponse{
		Resource: grpcutil.ResourceToProto(res),
	}, nil
}
