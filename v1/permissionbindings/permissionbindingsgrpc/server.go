package permissionbindingsgrpc

import (
	"context"

	"github.com/google/uuid"
	permissionbindingspb "github.com/ooqls/go-auth/gen/grpc/permissionbindings/v1"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ permissionbindingspb.PermissionBindingsServiceServer = (*Server)(nil)

type Server struct {
	permissionbindingspb.UnimplementedPermissionBindingsServiceServer
	reader permissionbindingsv1.Reader
	writer permissionbindingsv1.Writer
	l      *zap.Logger
}

func NewServer(reader permissionbindingsv1.Reader, writer permissionbindingsv1.Writer, l *zap.Logger) *Server {
	return &Server{
		reader: reader,
		writer: writer,
		l:      l,
	}
}

func (s *Server) AssignPermission(ctx context.Context, req *permissionbindingspb.AssignPermissionRequest) (*permissionbindingspb.AssignPermissionResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	roleIDs, err := grpcutil.ParseUUIDs(req.GetRoleIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role ids")
	}

	permIDs, err := grpcutil.ParseUUIDs(req.GetPermissionIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid permission ids")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	for _, roleID := range roleIDs {
		for _, permID := range permIDs {
			if err := s.writer.AssignPermission(lc, roleID, permID); err != nil {
				return nil, grpcutil.HandleError(err)
			}
		}
	}

	return &permissionbindingspb.AssignPermissionResponse{}, nil
}

func (s *Server) UnassignPermission(ctx context.Context, req *permissionbindingspb.UnassignPermissionRequest) (*permissionbindingspb.UnassignPermissionResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	roleIDs, err := grpcutil.ParseUUIDs(req.GetRoleIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role ids")
	}

	permIDs, err := grpcutil.ParseUUIDs(req.GetPermissionIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid permission ids")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	for _, roleID := range roleIDs {
		for _, permID := range permIDs {
			if err := s.writer.UnassignPermission(lc, roleID, permID); err != nil {
				return nil, grpcutil.HandleError(err)
			}
		}
	}

	return &permissionbindingspb.UnassignPermissionResponse{}, nil
}

func (s *Server) GetPermissionBindingsForRole(ctx context.Context, req *permissionbindingspb.GetPermissionBindingsForRoleRequest) (*permissionbindingspb.GetPermissionBindingsForRoleResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role id")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	bindings, err := s.reader.GetPermissionBindingsForRole(lc, roleID)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &permissionbindingspb.GetPermissionBindingsForRoleResponse{
		PermissionBindings: grpcutil.PermissionBindingsToProto(bindings),
	}, nil
}

func (s *Server) ListPermissionBindings(ctx context.Context, req *permissionbindingspb.ListPermissionBindingsRequest) (*permissionbindingspb.ListPermissionBindingsResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
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

	lc := contexts.NewLoggingContext(ctx, s.l)
	bindings, err := s.reader.GetPermissionsBindings(lc, page, pageSize)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &permissionbindingspb.ListPermissionBindingsResponse{
		PermissionBindings: grpcutil.PermissionBindingsToProto(bindings),
	}, nil
}
