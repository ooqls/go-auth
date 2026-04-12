package rolesgrpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rolespb "github.com/ooqls/go-auth/gen/grpc/roles/v1"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/roles"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ rolespb.RolesServiceServer = (*Server)(nil)

type Server struct {
	rolespb.UnimplementedRolesServiceServer
	service           roles.Service
	permissionService permissions.Service
	l                 *zap.Logger
}

func NewServer(
	svc roles.Service,
	permSvc permissions.Service,
	l *zap.Logger,
) *Server {
	return &Server{
		service:           svc,
		permissionService: permSvc,
		l:                 l,
	}
}

func (s *Server) GetAuthRoles(ctx context.Context, req *rolespb.GetAuthRolesRequest) (*rolespb.GetAuthRolesResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	page := int32(1)
	pageSize := int32(100)
	if req.Page != nil {
		page = req.GetPage()
	}
	if req.PageSize != nil {
		pageSize = req.GetPageSize()
	}

	roleList, err := s.service.ListRoles(authCtx, page, pageSize)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &rolespb.GetAuthRolesResponse{
		Roles: grpcutil.RolesToProto(roleList),
	}, nil
}

func (s *Server) GetAuthRole(ctx context.Context, req *rolespb.GetAuthRoleRequest) (*rolespb.GetAuthRoleResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role id")
	}

	role, err := s.service.GetRole(authCtx, id)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	if role == nil {
		return nil, status.Error(codes.NotFound, "role not found")
	}

	return &rolespb.GetAuthRoleResponse{
		Role: grpcutil.RoleToProto(role),
	}, nil
}

func (s *Server) CreateAuthRole(ctx context.Context, req *rolespb.CreateAuthRoleRequest) (*rolespb.CreateAuthRoleResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	roleId, err := s.service.CreateRole(authCtx, roles.CreateRoleParams{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Hierarchy:   req.GetHierarchy(),
	})
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &rolespb.CreateAuthRoleResponse{
		Id: roleId.String(),
	}, nil
}

func (s *Server) UpdateAuthRole(ctx context.Context, req *rolespb.UpdateAuthRoleRequest) (*rolespb.UpdateAuthRoleResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role id")
	}

	updateParams := roles.UpdateRoleParams{ID: id}
	for _, ev := range req.GetEvents() {
		switch ev.GetEvent() {
		case "update_name":
			if name, ok := ev.GetData()["name"]; ok {
				updateParams.Name = &name
			}
		case "update_description":
			if desc, ok := ev.GetData()["description"]; ok {
				updateParams.Description = &desc
			}
		case "update_hierarchy":
			if h, ok := ev.GetData()["hierarchy"]; ok {
				var parsed int64
				_, err := fmt.Sscanf(h, "%d", &parsed)
				if err != nil {
					return nil, status.Error(codes.InvalidArgument, "invalid hierarchy value")
				}
				h32 := int32(parsed)
				updateParams.Hierarchy = &h32
			}
		}
	}

	if err := s.service.UpdateRole(authCtx, updateParams); err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &rolespb.UpdateAuthRoleResponse{}, nil
}

func (s *Server) DeleteAuthRole(ctx context.Context, req *rolespb.DeleteAuthRoleRequest) (*rolespb.DeleteAuthRoleResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role id")
	}

	if err := s.service.DeleteRole(authCtx, id); err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &rolespb.DeleteAuthRoleResponse{}, nil
}

func (s *Server) GetAuthPermissions(ctx context.Context, req *rolespb.GetAuthPermissionsRequest) (*rolespb.GetAuthPermissionsResponse, error) {
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

	perms, err := s.permissionService.GetPermissions(authCtx, page, pageSize)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	protoPerms := grpcutil.PermissionsToProto(perms)

	return &rolespb.GetAuthPermissionsResponse{
		Permissions: protoPerms,
	}, nil
}

func (s *Server) CreateAuthPermission(ctx context.Context, req *rolespb.CreateAuthPermissionRequest) (*rolespb.CreateAuthPermissionResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	for _, p := range req.GetPermissions() {
		if err := s.permissionService.AddPermission(authCtx, p.GetResourceName(), p.GetResourceGroup(), p.GetResourceKind(), p.GetActions()); err != nil {
			return nil, grpcutil.HandleError(err)
		}
	}

	return &rolespb.CreateAuthPermissionResponse{}, nil
}
