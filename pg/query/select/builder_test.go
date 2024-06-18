package sel

import (
	"testing"

	"github.com/braumsmilk/go-auth/pg/query"
	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder(t *testing.T) {
	qb := QueryBuilder{}
	q, err := qb.Select([]string{"col1"}).From("table").Where(
		query.KeyValue{
			Key:   "key1",
			Value: 1,
		},
		query.KeyValue{
			Key:   "key2",
			Value: "2",
		}).Limit(10).Offset(5).Build()

	assert.Nilf(t, err, "should not get error when building query")
	assert.Truef(t, len(q) > 0, "should have gotten a query")

	qb = QueryBuilder{}
	_, err = qb.SelectAll().Where(query.KeyValue{Key: "k1", Value: "1"}).Build()
	assert.NotNilf(t, err, "should have gotten an error when no table is selected")
}
