package update

import (
	"testing"

	"github.com/braumsmilk/go-auth/pg/query"
)

func TestQueryBuilder(t *testing.T) {
	qb := QueryBuilder{}

	qb.Update("table").Set(query.KeyValue{Key: "key1", Value: "val2"})
}
