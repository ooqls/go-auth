package usersgrpc

import (
	"context"

	"github.com/google/uuid"
	userspb "github.com/ooqls/go-auth/gen/grpc/users/v1"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"github.com/ooqls/go-auth/v1/users"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ userspb.UsersServiceServer = (*Server)(nil)

type Server struct {
	userspb.UnimplementedUsersServiceServer
	service users.Service
	l       *zap.Logger
}

func NewServer(svc users.Service, l *zap.Logger) *Server {
	return &Server{
		service: svc,
		l:       l,
	}
}

func (s *Server) GetCurrentUser(ctx context.Context, _ *userspb.GetCurrentUserRequest) (*userspb.GetCurrentUserResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	user, err := s.service.GetUser(authCtx, authCtx.GetAuthedUser().Id)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &userspb.GetCurrentUserResponse{
		User: grpcutil.UserToProto(user),
	}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *userspb.CreateUserRequest) (*userspb.CreateUserResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if req.Key != nil && req.Salt != nil {
		user, err := s.service.CreateUserWithPassword(authCtx, req.GetEmail(), req.GetUsername(), req.GetKey(), req.GetSalt())
		if err != nil {
			return nil, grpcutil.HandleError(err)
		}
		return &userspb.CreateUserResponse{
			User: grpcutil.UserToProto(user),
		}, nil
	}

	user, pw, err := s.service.CreateUserWithRandomPassword(authCtx, req.GetEmail(), req.GetUsername())
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &userspb.CreateUserResponse{
		User:     grpcutil.UserToProto(user),
		Password: &pw,
	}, nil
}

func (s *Server) GetUserByID(ctx context.Context, req *userspb.GetUserByIDRequest) (*userspb.GetUserByIDResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.service.GetUser(authCtx, id)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &userspb.GetUserByIDResponse{
		User: grpcutil.UserToProto(user),
	}, nil
}

func (s *Server) UpdateUserByID(ctx context.Context, req *userspb.UpdateUserByIDRequest) (*userspb.UpdateUserByIDResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	for _, ev := range req.GetEvents() {
		switch ev.GetEvent() {
		case "email":
			emailVal := ev.GetData()["email"]
			if emailVal == "" {
				return nil, status.Error(codes.InvalidArgument, "email value required")
			}
			if err := s.service.UpdateUserEmail(authCtx, id, emailVal); err != nil {
				return nil, grpcutil.HandleError(err)
			}
		default:
			return nil, status.Errorf(codes.InvalidArgument, "invalid event: %s", ev.GetEvent())
		}
	}

	return &userspb.UpdateUserByIDResponse{}, nil
}

func (s *Server) DeleteUserByID(ctx context.Context, req *userspb.DeleteUserByIDRequest) (*userspb.DeleteUserByIDResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := s.service.DeleteUser(authCtx, id); err != nil {
		return nil, grpcutil.HandleError(err)
	}

	return &userspb.DeleteUserByIDResponse{}, nil
}

func (s *Server) GetUsers(ctx context.Context, req *userspb.GetUsersRequest) (*userspb.GetUsersResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	page := 0
	pageSize := 10
	if req.Page != nil {
		page = int(req.GetPage())
	}
	if req.PageSize != nil {
		pageSize = int(req.GetPageSize())
	}

	userList, err := s.service.GetUsers(authCtx, page, pageSize)
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	resp := &userspb.GetUsersResponse{}
	for i := range userList {
		resp.Users = append(resp.Users, grpcutil.UserToProto(&userList[i]))
	}

	return resp, nil
}

func (s *Server) GetUserByUsername(ctx context.Context, req *userspb.GetUserByUsernameRequest) (*userspb.GetUserByUsernameResponse, error) {
	authCtx, ok := grpcutil.AuthorizationFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username required")
	}

	user, err := s.service.GetUserByUsername(authCtx, req.GetUsername())
	if err != nil {
		return nil, grpcutil.HandleError(err)
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &userspb.GetUserByUsernameResponse{
		User: grpcutil.UserToProto(user),
	}, nil
}
