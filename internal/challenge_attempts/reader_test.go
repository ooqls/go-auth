package challengeattempts

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/elasticsearch"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	containers.StartElasticsearch(ctx, containers.WithLogging())
	os.Exit(m.Run())
}

func TestReaderWriter(t *testing.T) {
	ctx := context.Background()
	assert.Nilf(t, elasticsearch.InitDefault(), "should not failt o init elasticsearch")
	elastic := elasticsearch.Get()
	writer := NewDocumentWriter(elastic)
	reader := NewDocumentReader(elastic)

	challengeAttempt := ChallengeAttempt{
		ID:          uuid.New(),
		ChallengeID: uuid.New(),
		UserID:      uuid.New(),
		Success:     true,
		CreatedAt:   time.Now(),
	}
	err := writer.CreateChallengeAttempt(ctx, challengeAttempt)
	assert.Nil(t, err)
	attempts, err := reader.GetChallengeAttempts(ctx, challengeAttempt.UserID.String(), 1, 10)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(attempts))
}
