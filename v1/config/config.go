package config

import (
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/crypto/jwt"
)

func GetStandaloneAppConfig(port int) *app.AppConfig {
	return &app.AppConfig{
		ServerConfig: app.ServerConfig{
			Port: port,
		},
		SQLFiles: app.SQLFilesConfig{
			Enabled:      true,
			SQLPackage:   app.SQLXPackage,
			SQLFilesDirs: []string{"./migrations/"},
		},
		Gin: app.GinConfig{
			Enabled: true,
			Port:    8080,
			Cors: app.CorsConfig{
				Enabled:             true,
				AllowMethods:        []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowWildcard:       true,
				AllowCredentials:    true,
				AllowAllOrigins:     true,
				AllowPrivateNetwork: true,
				ExposeHeaders:       []string{"Content-Type", "Authorization"},
				Headers:             []string{"Content-Type", "Authorization"},
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
		DocsConfig: app.DocsConfig{
			Enabled:     true,
			DocsDir:     "./docs/",
			DocsApiPath: "/api/v1/docs",
		},
	}
}
