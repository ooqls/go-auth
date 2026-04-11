package authenticationgrpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/ooqls/getset/crypto/crypto"
	authpb "github.com/ooqls/go-auth/gen/grpc/auth/v1"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/v1/authentication"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ authpb.AuthServiceServer = (*Server)(nil)

type Server struct {
	authpb.UnimplementedAuthServiceServer
	service authentication.Service
	l       *zap.Logger
}

func NewServer(svc authentication.Service, l *zap.Logger) *Server {
	return &Server{
		service: svc,
		l:       l,
	}
}

func (s *Server) IsAuthenticated(ctx context.Context, req *authpb.IsAuthenticatedRequest) (*authpb.IsAuthenticatedResponse, error) {
	token := req.GetToken()
	if token == "" {
		// Try to get from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if tokens := md.Get("authorization"); len(tokens) > 0 {
				token = tokens[0]
			}
		}
	}

	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "token required")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	claims, err := s.service.IsAuthenticated(lc, token)
	if err != nil {
		s.l.Error("failed to authenticate token", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	return &authpb.IsAuthenticatedResponse{
		UserId: claims.UserID.String(),
	}, nil
}

func (s *Server) VerifyKey(ctx context.Context, req *authpb.VerifyKeyRequest) (*authpb.VerifyKeyResponse, error) {
	err := crypto.VerifyGCMAESKey(req.GetKey())
	if err != nil {
		return &authpb.VerifyKeyResponse{Success: false}, nil
	}
	return &authpb.VerifyKeyResponse{Success: true}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token required")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	resp, err := s.service.AuthenticateWithToken(lc, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	return &authpb.RefreshTokenResponse{
		UserId:    resp.UserId.String(),
		AuthToken: resp.AuthToken,
	}, nil
}

func (s *Server) LoginChallenge(ctx context.Context, req *authpb.LoginChallengeRequest) (*authpb.LoginChallengeResponse, error) {
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username required")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	challenge, err := s.service.IssueChallenge(lc, req.GetUsername())
	if err != nil {
		s.l.Error("failed to issue challenge", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to issue challenge")
	}

	return &authpb.LoginChallengeResponse{
		Id:        challenge.ID.String(),
		Challenge: challenge.Challenge,
		Salt:      challenge.User.Salt,
	}, nil
}

func (s *Server) AnswerLoginChallenge(ctx context.Context, req *authpb.AnswerLoginChallengeRequest) (*authpb.AnswerLoginChallengeResponse, error) {
	challengeID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid challenge id")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	result, err := s.service.ValidateChallenge(lc, challengeID, req.GetChallenge())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	return &authpb.AnswerLoginChallengeResponse{
		AuthToken:    result.Okey,
		RefreshToken: result.Rkey,
		UserId:       result.UserID,
	}, nil
}

func (s *Server) Logout(ctx context.Context, _ *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	// gRPC is stateless — logout is a no-op on the server side.
	// The client should discard its tokens.
	return &authpb.LogoutResponse{}, nil
}

func (s *Server) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	if req.GetUsername() == "" || req.GetEmail() == "" || len(req.GetKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "username, email, and key are required")
	}

	lc := contexts.NewLoggingContext(ctx, s.l)
	result, err := s.service.Register(lc, authenticationv1.RegistrationParam{
		Username: req.GetUsername(),
		Key:      req.GetKey(),
		Email:    req.GetEmail(),
	})
	if err != nil {
		s.l.Error("failed to register user", zap.Error(err))
		return nil, grpcutil.HandleError(err)
	}

	return &authpb.RegisterResponse{
		AuthToken:    result.Okey,
		RefreshToken: result.Rkey,
		UserId:       result.UserID,
	}, nil
}
