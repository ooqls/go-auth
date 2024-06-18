package sel

import (
	"fmt"
	"strings"

	"github.com/braumsmilk/go-auth/pg/query"
)

type QueryBuilder struct {
	table   *string
	columns []string
	limit   *int
	offset  *int
	kv      []query.KeyValue
}


func (qb *QueryBuilder) SelectAll(kv ...query.KeyValue) *QueryBuilder {
	return qb.Select([]string{"*"})
}

func (qb *QueryBuilder) Select(cols []string) *QueryBuilder {
	qb.columns = cols
	return qb
}

func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.table = &table
	return qb
}

func (qb *QueryBuilder) Where(kv ...query.KeyValue) *QueryBuilder {
	qb.kv = kv
	return qb
}

func (qb *QueryBuilder) Limit(l int) *QueryBuilder {
	qb.limit = &l
	return qb
}

func (qb *QueryBuilder) Offset(off int) *QueryBuilder {
	qb.offset = &off
	return qb
}

func (qb *QueryBuilder) Build() (string, error) {
	if qb.table == nil {
		return "", query.MissingTableErr
	}

	q := fmt.Sprintf("SELECT %s FROM %s", strings.Join(qb.columns, ","), *qb.table)
	if len(qb.kv) > 0 {
		q += " WHERE "
		for _, kv := range qb.kv {
			q = q + fmt.Sprintf("%s AND ", kv.String())
		}
		q = strings.TrimRight(q, " AND ")
	}

	if qb.offset != nil {
		q = q + fmt.Sprintf(" OFFSET %d", *qb.offset)
	}

	if qb.limit != nil {
		q = q + fmt.Sprintf(" LIMIT %d", *qb.limit)
	}

	return q, nil
}
