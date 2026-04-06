package challengeattempts

import (
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

//go:generate go run github.com/golang/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks -mock_names=Reader=MockReader
type Reader interface {
	GetChallengeAttempts(ctx context.Context, userID string) ([]ChallengeAttempt, error)
	GetFailedAttempts(ctx context.Context, userID string, minutes int) ([]ChallengeAttempt, error)
}

type DocumentReader struct {
	cli *elasticsearch.TypedClient
	l   *zap.Logger
}

func NewDocumentReader(cli *elasticsearch.TypedClient) *DocumentReader {
	return &DocumentReader{
		cli: cli,
		l:   log.NewLogger("DocumentReader"),
	}
}

func (r *DocumentReader) GetChallengeAttempts(ctx context.Context, userID string, page, pageSize int) ([]ChallengeAttempt, error) {

	// Create an Elasticsearch query to find documents matching userID and paginate results.
	// pageSize and page are parameters to this method.

	q := &types.Query{
		Bool: &types.BoolQuery{
			Must: []types.Query{
				{
					Match: map[string]types.MatchQuery{
						"user_id": {
							Query: userID,
						},
					},
				},
			},
		},
	}

	s := r.cli.API.Search()
	s.Index("challengeattemptsv1")
	s.Query(q)
	s.From((page - 1) * pageSize)
	s.Size(pageSize)

	res, err := s.Do(ctx)
	if err != nil {
		return nil, err
	}

	attempts := make([]ChallengeAttempt, len(res.Hits.Hits))
	for i, hit := range res.Hits.Hits {
		var attempt ChallengeAttempt
		if err := json.Unmarshal(hit.Source_, &attempt); err != nil {
			return nil, err
		}
		attempts[i] = attempt
	}
	return attempts, nil
}
