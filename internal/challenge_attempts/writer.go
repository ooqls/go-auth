package challengeattempts

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

//go:generate go run github.com/golang/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks -mock_names=Writer=MockWriter
type Writer interface {
	CreateChallengeAttempt(ctx context.Context, challengeAttempt ChallengeAttempt) error
}

type DocumentWriter struct {
	cli *elasticsearch.TypedClient
	l   *zap.Logger
}

func NewDocumentWriter(cli *elasticsearch.TypedClient) *DocumentWriter {
	return &DocumentWriter{
		cli: cli,
		l:   log.NewLogger("DocumentWriter"),
	}
}

func (w *DocumentWriter) CreateChallengeAttempt(ctx context.Context, challengeAttempt ChallengeAttempt) error {
	_, err := w.cli.Index("challengeattemptsv1").Document(challengeAttempt).Refresh(refresh.True).Do(ctx)
	return err
}
