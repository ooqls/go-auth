package insert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder(t *testing.T) {
	qb := QueryBuilder{}
	q, err := qb.Insert("table", []string{"col1", "col2"}).Build()
	assert.Nilf(t, err, "should not get error when building query")
	assert.Truef(t, len(q) > 0, "query should not be empty")

	qb = QueryBuilder{}
	_, err = qb.Build()
	assert.NotNilf(t, err, "should get error when no table specified")
}
