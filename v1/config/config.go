package config

import (
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/go-auth/internal/schema"
)

func GetStandaloneAppConfig(port int) *app.AppConfig {
	return &app.AppConfig{
		ServerConfig: app.ServerConfig{
			Port: port,
		},
		SQLFiles: app.SQLFilesConfig{
			Enabled:          true,
			SQLPackage:       app.SQLXPackage,
			SQLFilesDirs:     []string{"./migrations/"},
			CreateTableStmts: []string{schema.Migration20260114144208Auth, schema.Migration20260309021330UserHighestRoleView},
		},

		Gin: app.GinConfig{
			Enabled: true,
			Port:    port,
			Cors: app.CorsConfig{
				Enabled:             true,
				AllowMethods:        []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowCredentials:    true,
				AllowOrigins:        []string{"http://localhost:5173", "http://localhost:8080", "http://localhost:3000"},
				AllowWildcard:       true,
				AllowPrivateNetwork: true,
				ExposeHeaders:       []string{"Origin", "Content-Type", "Authorization", "X-User-Id"},
				Headers:             []string{"Origin", "Content-Type", "Authorization", "X-User-Id"},
				MaxAge:              12,
			},
		},
		LoggingAPI: app.LoggingAPIConfig{
			Enabled: false,
		},
		JWT: app.JWTConfig{
			Enabled: true,
			TokenConfigurations: []jwt.TokenConfiguration{
				{
					Audience:                []string{"auth"},
					Issuer:                  app.AuthIssuer,
					ValidityDurationSeconds: 3600,
				},
				{
					Audience:                []string{"refresh"},
					Issuer:                  app.RefreshIssuer,
					ValidityDurationSeconds: 3600,
				},
			},
		},
		Cache: app.CacheConfig{
			Enabled:   true,
			CacheType: "valkey",
		},
		DocsConfig: app.DocsConfig{
			Enabled:     true,
			DocsApiPath: "/api/docs",
		},
	}
}
