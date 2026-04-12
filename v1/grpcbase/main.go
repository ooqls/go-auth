package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/log"
	authpb "github.com/ooqls/go-auth/gen/grpc/auth/v1"
	permissionbindingspb "github.com/ooqls/go-auth/gen/grpc/permissionbindings/v1"
	permissionspb "github.com/ooqls/go-auth/gen/grpc/permissions/v1"
	resourcespb "github.com/ooqls/go-auth/gen/grpc/resources/v1"
	rolebindingspb "github.com/ooqls/go-auth/gen/grpc/rolebindings/v1"
	rolespb "github.com/ooqls/go-auth/gen/grpc/roles/v1"
	userspb "github.com/ooqls/go-auth/gen/grpc/users/v1"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/schema"
	"github.com/ooqls/go-auth/v1/authentication"
	"github.com/ooqls/go-auth/v1/authentication/authenticationgrpc"
	"github.com/ooqls/go-auth/v1/config"
	grpcutil "github.com/ooqls/go-auth/v1/grpc"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/permissions/permissionsgrpc"
	"github.com/ooqls/go-auth/v1/permissionbindings/permissionbindingsgrpc"
	"github.com/ooqls/go-auth/v1/resources/resourcesgrpc"
	"github.com/ooqls/go-auth/v1/rolebindings"
	"github.com/ooqls/go-auth/v1/rolebindings/rolebindingsgrpc"
	"github.com/ooqls/go-auth/v1/roles"
	"github.com/ooqls/go-auth/v1/roles/rolesgrpc"
	"github.com/ooqls/go-auth/v1/users"
	"github.com/ooqls/go-auth/v1/users/usersgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	appConfigPath   string
	standalone      bool
	testEnvironment bool
	grpcPort        int
	api             string
)

func init() {
	flag.StringVar(&appConfigPath, "app-config", "", "path to app config")
	flag.BoolVar(&standalone, "standalone", false, "run in standalone mode")
	flag.BoolVar(&testEnvironment, "test-environment", false, "run in test environment")
	flag.IntVar(&grpcPort, "grpc-port", 9090, "gRPC server port")
	flag.StringVar(&api, "api", "all", "api to serve: all, authentication, users, roles, permissions, permissionbindings, resources, rolebindings")
}

func main() {
	flag.Parse()
	l := log.NewLogger("grpc-" + api)

	// Build the app to initialize infrastructure (DB, cache, etc.)
	var appConfig *app.AppConfig
	if standalone {
		appConfig = config.GetStandaloneAppConfig(grpcPort)
	} else {
		var err error
		appConfig, err = app.LoadConfig(appConfigPath)
		if err != nil {
			l.Fatal("failed to load app config", zap.Error(err))
		}
	}

	application := app.New("grpc-server", app.WithConfig(appConfig))
	if testEnvironment || standalone {
		application.WithTestEnvironment(app.TestEnvironment{
			Redis:         false,
			Postgres:      true,
			Elasticsearch: true,
		})
	}

	// Channel for the gRPC server so OnRunning can signal it
	grpcServerCh := make(chan *grpc.Server, 1)
	errCh := make(chan error, 1)

	application.OnRunning(func(ctx *app.AppContext) error {
		cacheFactory, _ := ctx.CacheFactory()
		factory := datav1.NewFactory(*pgx.GetPGX(), cacheFactory)

		// Create gRPC server with auth interceptor
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(grpcutil.UnaryAuthInterceptor(factory, l)),
			grpc.StreamInterceptor(grpcutil.StreamAuthInterceptor(factory, l)),
		)

		// Register services based on --api flag
		switch api {
		case "authentication":
			registerAuthentication(ctx, grpcServer, factory, l)
		case "users":
			registerUsers(grpcServer, factory, l)
		case "roles":
			registerRoles(grpcServer, factory, l)
		case "permissions":
			registerPermissions(grpcServer, factory, l)
		case "permissionbindings":
			registerPermissionBindings(grpcServer, factory, l)
		case "resources":
			registerResources(grpcServer, factory, l)
		case "rolebindings":
			registerRoleBindings(grpcServer, factory, l)
		default: // "all"
			registerAuthentication(ctx, grpcServer, factory, l)
			registerUsers(grpcServer, factory, l)
			registerRoles(grpcServer, factory, l)
			registerPermissions(grpcServer, factory, l)
			registerPermissionBindings(grpcServer, factory, l)
			registerResources(grpcServer, factory, l)
			registerRoleBindings(grpcServer, factory, l)
		}

		// Enable reflection for grpcurl/debugging
		reflection.Register(grpcServer)

		grpcServerCh <- grpcServer

		// Start listening
		addr := fmt.Sprintf(":%d", grpcPort)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			errCh <- fmt.Errorf("failed to listen on %s: %w", addr, err)
			return err
		}

		l.Info("gRPC server starting", zap.String("address", addr), zap.String("api", api))
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- err
			return err
		}
		return nil
	})

	// Run the app (initializes DB, cache, etc.) in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		application.Run(ctx)
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		l.Info("received signal, shutting down", zap.String("signal", sig.String()))
		select {
		case srv := <-grpcServerCh:
			srv.GracefulStop()
		default:
		}
	case err := <-errCh:
		l.Error("gRPC server error", zap.Error(err))
	}

	cancel()
}

func registerAuthentication(ctx *app.AppContext, grpcServer *grpc.Server, factory datav1.Factory, l *zap.Logger) {
	refreshCfg, ok := ctx.RefreshIssuerConfig()
	if !ok {
		l.Warn("refresh issuer config not found, skipping authentication service registration")
		return
	}
	authCfg, ok := ctx.AuthIssuerConfig()
	if !ok {
		l.Warn("auth issuer config not found, skipping authentication service registration")
		return
	}

	refreshIssuer := jwt.NewJwtTokenIssuer[claims.UserClaims](refreshCfg, keys.JWT())
	authenticationv1.InitRefreshIssuer(refreshIssuer)

	authenticationIssuer := jwt.NewJwtTokenIssuer[claims.UserClaims](authCfg, keys.JWT())
	authenticationv1.InitAuthenticationIssuer(authenticationIssuer)

	cacheFactory, _ := ctx.CacheFactory()
	userIssuer := authenticationv1.NewUserTokenIssuer(authenticationIssuer, refreshIssuer)
	challenger := authenticationv1.NewChallengerV1(cacheFactory.NewStore("challenges", 15*time.Minute))
	authenticator := authenticationv1.NewAuthenticatorV1(
		userIssuer,
		cacheFactory,
		challenger,
		[]string{"auth"},
	)

	userService := users.NewServiceImpl(factory)
	authService := authentication.NewServiceImpl(authenticator, userService)
	authpb.RegisterAuthServiceServer(grpcServer, authenticationgrpc.NewServer(authService, l))
	l.Info("registered gRPC AuthService")
}

func registerUsers(grpcServer *grpc.Server, factory datav1.Factory, l *zap.Logger) {
	userService := users.NewServiceImpl(factory)
	userspb.RegisterUsersServiceServer(grpcServer, usersgrpc.NewServer(userService, l))
	l.Info("registered gRPC UsersService")
}

func registerRoles(grpcServer *grpc.Server, factory datav1.Factory, l *zap.Logger) {
	rolesService := roles.NewServiceImpl(factory)
	permissionService := permissions.NewServiceImpl(factory)
	rolespb.RegisterRolesServiceServer(grpcServer, rolesgrpc.NewServer(rolesService, permissionService, l))
	l.Info("registered gRPC RolesService")
}

func registerPermissions(grpcServer *grpc.Server, factory datav1.Factory, l *zap.Logger) {
	permissionService := permissions.NewServiceImpl(factory)
	permissionspb.RegisterPermissionsServiceServer(grpcServer, permissionsgrpc.NewServer(permissionService, l))
	l.Info("registered gRPC PermissionsService")
}

func registerPermissionBindings(grpcServer *grpc.Server, factory datav1.Factory, l *zap.Logger) {
	pbReader := factory.NewPermissionBindingReader()
	pbWriter := factory.NewPermissionBindingWriter()
	permissionbindingspb.RegisterPermissionBindingsServiceServer(grpcServer, permissionbindingsgrpc.NewServer(pbReader, pbWriter, l))
	l.Info("registered gRPC PermissionBindingsService")
}

func registerResources(grpcServer *grpc.Server, factory datav1.Factory, l *zap.Logger) {
	resourcespb.RegisterResourcesServiceServer(grpcServer, resourcesgrpc.NewServer(factory, l))
	l.Info("registered gRPC ResourcesService")
}

func registerRoleBindings(grpcServer *grpc.Server, factory datav1.Factory, l *zap.Logger) {
	auth := authorizationv1.NewAuthorizerImpl(factory)
	rbReader := factory.NewRoleBindingsReader()
	rbWriter := factory.NewRoleBindingsWriter()
	rbService := rolebindings.NewServiceImpl(auth, rbReader, rbWriter)
	rolebindingspb.RegisterRoleBindingsServiceServer(grpcServer, rolebindingsgrpc.NewServer(rbService, l))
	l.Info("registered gRPC RoleBindingsService")
}

// Ensure schema is available for standalone/test mode.
var _ = schema.GetSchemaMigrations
