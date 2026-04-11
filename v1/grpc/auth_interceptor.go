package grpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/usersv1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authContextKey struct{}

// AuthorizationFromContext extracts the authorizationv1.Context from a Go context.
func AuthorizationFromContext(ctx context.Context) (*authorizationv1.Context, bool) {
	authCtx, ok := ctx.Value(authContextKey{}).(authorizationv1.Context)
	if !ok {
		return nil, false
	}
	return &authCtx, true
}

// UnaryAuthInterceptor creates a gRPC unary server interceptor that
// extracts the user identity from the "x-user-id" metadata header
// (set by a trusted upstream proxy) and populates the authorization context.
func UnaryAuthInterceptor(factory datav1.Factory, l *zap.Logger) grpc.UnaryServerInterceptor {
	userReader := factory.NewUserReader()
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		newCtx, err := buildAuthContext(ctx, userReader, l)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// StreamAuthInterceptor creates a gRPC stream server interceptor that
// extracts the user identity from the "x-user-id" metadata header.
func StreamAuthInterceptor(factory datav1.Factory, l *zap.Logger) grpc.StreamServerInterceptor {
	userReader := factory.NewUserReader()
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		newCtx, err := buildAuthContext(ctx, userReader, l)
		if err != nil {
			return err
		}
		wrapped := &wrappedServerStream{ServerStream: ss, ctx: newCtx}
		return handler(srv, wrapped)
	}
}

func buildAuthContext(ctx context.Context, userReader usersv1.Reader, l *zap.Logger) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		authCtx := authorizationv1.NewUnauthenticatedContext(ctx)
		return context.WithValue(ctx, authContextKey{}, authCtx), nil
	}

	// Try x-user-id header first (trusted proxy mode, matching HTTP TrustedMiddleware)
	userIDs := md.Get("x-user-id")
	if len(userIDs) == 0 {
		// Also check authorization header for direct token usage
		authHeaders := md.Get("authorization")
		if len(authHeaders) > 0 {
			token := strings.TrimPrefix(authHeaders[0], "Bearer ")
			token = strings.TrimPrefix(token, "bearer ")
			// Store token in context for services that need it
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
		}
		authCtx := authorizationv1.NewUnauthenticatedContext(ctx)
		return context.WithValue(ctx, authContextKey{}, authCtx), nil
	}

	userID := userIDs[0]
	l.Info("gRPC auth: user ID from metadata", zap.String("userID", userID))

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		l.Error("failed to parse user ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid user ID")
	}

	lc := contexts.NewLoggingContext(ctx, l)
	userObj, err := userReader.GetUser(lc, userIDUUID)
	if err != nil {
		l.Error("failed to get user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if userObj == nil {
		l.Info("user not found", zap.String("userID", userID))
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	authCtx := authorizationv1.NewAuthorizationContext(*userObj)
	return context.WithValue(ctx, authContextKey{}, authCtx), nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
