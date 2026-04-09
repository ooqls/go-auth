package main

import (
	"os"

	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/registry"
	"go.yaml.in/yaml/v2"
)

func main() {
	reg := registry.Registry{
		Redis: &registry.Database{
			Server: registry.Server{
				Name: "redis",
				Host: "localhost",
				Port: 6379,
				Auth: registry.Auth{
					Username:     "",
					PasswordFile: "password.txt",
				},
				TLS: &registry.TLSConfig{
					Enabled:  true,
					CaPath:   "/app/tls/ca.pem",
					CertPath: "/app/tls/server.pem",
					KeyPath:  "/app/tls/server.key.pem",
				},
			},
			Database: "1",
		},
		Postgres: &registry.Database{
			Server: registry.Server{
				Name: "postgres",
				Host: "localhost",
				Port: 5432,
				Auth: registry.Auth{
					Username:     "",
					PasswordFile: "password.txt",
				},
				TLS: &registry.TLSConfig{
					Enabled:  true,
					CaPath:   "/app/tls/ca.pem",
					CertPath: "/app/tls/server.pem",
					KeyPath:  "/app/tls/server.key.pem",
				},
			},
			Database: "appdb",
		},
	}

	regYaml, err := yaml.Marshal(reg)
	if err != nil {
		panic(err)
	}
	os.WriteFile("./registry.yaml", regYaml, 0644)

	appConfig := app.AppConfig{
		LoggingAPI: app.LoggingAPIConfig{
			Enabled: false,
		},
		SQLFiles: app.SQLFilesConfig{
			Enabled:      true,
			SQLPackage:   app.SQLXPackage,
			SQLFilesDirs: []string{"/app/migrations/"},
		},
		Gin: app.GinConfig{
			Enabled: true,
			Port:    8080,
		},
		DocsConfig: app.DocsConfig{
			Enabled:     true,
			DocsDir:     "/app/docs/",
			DocsApiPath: "/api/v1/docs",
		},
		JWT: app.JWTConfig{
			Enabled: true,
			TokenConfigurations: []jwt.TokenConfiguration{
				{
					Audience: []string{"auth"},
					Issuer:   app.AuthIssuer,
				},
			},
			RSAKeyPath:    "/app/jwt/jwt.key.pem",
			RSAPubKeyPath: "/app/jwt/jwt.pub.key.pem",
		},
		Registry: app.RegistryConfig{
			Enabled: true,
			Path:    "/app/registry/registry.yaml",
		},
		TLS: app.TLSConfig{
			Enabled:  true,
			CertFile: "/app/tls/server.pem",
			KeyFile:  "/app/tls/server.key.pem",
		},
	}
	appYaml, err := yaml.Marshal(appConfig)
	if err != nil {
		panic(err)
	}
	os.WriteFile("./config.yaml", appYaml, 0644)
}
