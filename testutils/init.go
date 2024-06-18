package testutils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/braumsmilk/go-auth/init/seed"
	"github.com/braumsmilk/go-registry"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func InitRedis() testcontainers.Container {
	ctx := context.Background()
	c := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379"},
		WaitingFor:   &wait.LogStrategy{Log: "Ready to accept connections"},
		Env: map[string]string{
			"REDIS_PASSWORD": "password",
		},
	}

	gc := testcontainers.GenericContainerRequest{
		ContainerRequest: c,
		Started:          true,
	}

	container, err := testcontainers.GenericContainer(ctx, gc)
	if err != nil {
		panic(fmt.Errorf("failed to start redis container: %v", err))
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		panic(fmt.Errorf("failed to get mapped redis port: %v", err))
	}

	log.Printf("redis should be running at localhost:%d", port.Int())
	time.Sleep(time.Second * 5)

	registry.Set(registry.Registry{
		Redis: &registry.Server{
			Host: "localhost",
			Port: port.Int(),
			Auth: registry.Auth{
				Enabled:  true,
				Password: "password",
			},
		},
	})

	return container
}

func InitPostgres() testcontainers.Container {
	ctx := context.Background()
	c := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432"},
		Env: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "user100",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: &wait.LogStrategy{Log: "database system is ready to accept connections"},
	}

	gc := testcontainers.GenericContainerRequest{
		ContainerRequest: c,
		Started:          true,
	}
	container, err := testcontainers.GenericContainer(ctx, gc)
	if err != nil {
		panic(fmt.Errorf("failed to start postgres container: %v", err))
	}
	// host, err := container.Host(ctx)
	// if err != nil {
	// 	panic(fmt.Errorf("failed to get host ip from container: %v", err))
	// }

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		panic(fmt.Errorf("failed to get mapped postgres port: %v", err))
	}

	log.Printf("postgres should be running at localhost:%d", port.Int())
	time.Sleep(time.Second * 5)

	registry.Set(registry.Registry{
		Postgres: &registry.Server{
			Host: "localhost",
			Port: port.Int(),
			Auth: registry.Auth{
				Enabled:  true,
				Username: "user",
				Password: "user100",
			},
		},
	})

	seed.SeedPostgresDatabase()

	return container
}
