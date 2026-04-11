package rolebindingsgrpc

import (
	"context"

	"github.com/google/uuid"
	rolebindingspb "github.com/ooqls/go-auth/gen/grpc/rolebindings/v1"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"github.com/ooqls/go-auth/v1/rolebindings"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ rolebindingspb.RoleBindingsServiceServer = (*Server)(nil)

type Server struct {
	rolebindingspb.UnimplementedRoleBindingsServiceServer
	rb rolebindings.Service
	l  *zap.Logger
}

func NewServer(rb rolebindings.Service, l *zap.Logger) *Server {
	return &Server{rb: rb, l: l}
}

func (s *Server) GetRoleBindingsForUser(ctx context.Context, req *rolebindingspb.GetRoleBindingsForUserRequest) (*rolebindingspb.GetRoleBindingsForUserResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	bindings, err := s.rb.GetRoleBindingsForUser(authCtx, userID)
	if err != nil {
		s.l.Error("failed to get role bindings for user", zap.Error(err), zap.String("userID", req.GetUserId()))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &rolebindingspb.GetRoleBindingsForUserResponse{
		RoleBindings: grpcutil.RoleBindingsToProto(bindings),
	}, nil
}

func (s *Server) AssignRole(ctx context.Context, req *rolebindingspb.AssignRoleRequest) (*rolebindingspb.AssignRoleResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	userIDs, err := grpcutil.ParseUUIDs(req.GetUserIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ids")
	}

	roleIDs, err := grpcutil.ParseUUIDs(req.GetRoleIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role ids")
	}

	for _, userID := range userIDs {
		for _, roleID := range roleIDs {
			if err := s.rb.AssignRoleToUser(authCtx, userID, roleID); err != nil {
				s.l.Error("failed to assign role to user", zap.Error(err))
				return nil, status.Error(codes.Internal, "internal server error")
			}
		}
	}

	return &rolebindingspb.AssignRoleResponse{}, nil
}

func (s *Server) UnassignRole(ctx context.Context, req *rolebindingspb.UnassignRoleRequest) (*rolebindingspb.UnassignRoleResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	userIDs, err := grpcutil.ParseUUIDs(req.GetUserIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ids")
	}

	roleIDs, err := grpcutil.ParseUUIDs(req.GetRoleIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role ids")
	}

	for _, userID := range userIDs {
		for _, roleID := range roleIDs {
			if err := s.rb.UnassignRoleFromUser(authCtx, userID, roleID); err != nil {
				s.l.Error("failed to unassign role from user", zap.Error(err))
				return nil, status.Error(codes.Internal, "internal server error")
			}
		}
	}

	return &rolebindingspb.UnassignRoleResponse{}, nil
}

func (s *Server) GetAllRoleBindings(ctx context.Context, req *rolebindingspb.GetAllRoleBindingsRequest) (*rolebindingspb.GetAllRoleBindingsResponse, error) {
	_, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// TODO: Implement list all role bindings with pagination
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
