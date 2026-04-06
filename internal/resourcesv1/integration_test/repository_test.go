package integrationTest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/go-auth/internal/testutils"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	timeout := time.Second * 30
	c := containers.StartPostgres(ctx)
	defer c.Stop(ctx, &timeout)
	testutils.SeedDatabase()
	exitCode := m.Run()
	os.Exit(exitCode)
}
