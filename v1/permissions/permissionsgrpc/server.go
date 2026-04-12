package permissionsgrpc

import (
	"context"

	permissionspb "github.com/ooqls/go-auth/gen/grpc/permissions/v1"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"github.com/ooqls/go-auth/v1/permissions"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ permissionspb.PermissionsServiceServer = (*Server)(nil)

type Server struct {
	permissionspb.UnimplementedPermissionsServiceServer
	service permissions.Service
	l       *zap.Logger
}

func NewServer(svc permissions.Service, l *zap.Logger) *Server {
	return &Server{
		service: svc,
		l:       l,
	}
}

func (s *Server) ListPermissions(ctx context.Context, req *permissionspb.ListPermissionsRequest) (*permissionspb.ListPermissionsResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	page := 1
	pageSize := 100
	if req.Page != nil {
		page = int(req.GetPage())
	}
	if req.PageSize != nil {
		pageSize = int(req.GetPageSize())
	}

	perms, err := s.service.GetPermissions(authCtx, page, pageSize)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &permissionspb.ListPermissionsResponse{
		Permissions: grpcutil.PermissionsToProto(perms),
	}, nil
}

func (s *Server) CreatePermission(ctx context.Context, req *permissionspb.CreatePermissionRequest) (*permissionspb.CreatePermissionResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	perm := req.GetPermission()
	if perm == nil {
		return nil, status.Error(codes.InvalidArgument, "permission required")
	}

	if err := s.service.AddPermission(authCtx, perm.GetResourceName(), perm.GetResourceGroup(), perm.GetResourceKind(), perm.GetActions()); err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &permissionspb.CreatePermissionResponse{}, nil
}

func (s *Server) DeletePermission(ctx context.Context, req *permissionspb.DeletePermissionRequest) (*permissionspb.DeletePermissionResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if err := s.service.DeletePermission(authCtx, req.GetName(), req.GetGroup(), req.GetKind()); err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &permissionspb.DeletePermissionResponse{}, nil
}
