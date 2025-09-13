package authv1

import (
	"context"
	"testing"

	"github.com/ooqls/go-db/elasticsearch"
	"github.com/ooqls/go-db/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAuditWriter(t *testing.T) {
	ctx := context.Background()
	c := testutils.StartElasticsearch(ctx, testutils.WithLogging())
	defer c.Terminate(ctx)

	err := elasticsearch.InitDefault()
	assert.Nilf(t, err, "failed to initialize elasticsearch client: %s", err)
	cli := elasticsearch.Get()
	// Create a new audit writer
	writer := NewElasticsearchAuditWriter(cli, "audit")

	// Create a new audit
	audit := Audit{
		Version:     1,
		UserID:      "test-user",
		ResourceORN: "test-resource",
		Action:      "test-action",
	}

	// Write the audit to Elasticsearch
	err = writer.CreateAudit(context.Background(), audit)
	if err != nil {
		t.Fatalf("Error writing audit to Elasticsearch: %s", err)
	}
}
