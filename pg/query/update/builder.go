package update

import (
	"fmt"
	"strings"

	"github.com/braumsmilk/go-auth/pg/query"
)

type QueryBuilder struct {
	table      *string
	whereKv    []query.KeyValue
	setKv      []query.KeyValue
	returnVals []string
	updateAll  bool
}

func (qb *QueryBuilder) Update(table string) *QueryBuilder {
	qb.table = &table
	return qb
}

func (qb *QueryBuilder) Set(kv ...query.KeyValue) *QueryBuilder {
	qb.setKv = kv
	return qb
}

func (qb *QueryBuilder) Where(kv ...query.KeyValue) *QueryBuilder {
	qb.whereKv = kv
	return qb
}

func (qb *QueryBuilder) All() *QueryBuilder {
	qb.updateAll = true
	return qb
}

func (qb *QueryBuilder) Returning(r []string) *QueryBuilder {
	qb.returnVals = r
	return qb
}

func (qb *QueryBuilder) Build() (string, error) {
	if qb.table == nil {
		return "", query.MissingTableErr
	}

	if qb.setKv == nil || len(qb.setKv) == 0 {
		return "", query.MissingSetValuesErr
	}

	if !qb.updateAll && (qb.whereKv == nil || len(qb.whereKv) == 0) {
		return "", query.MissingWhereClauseErr
	}

	q := fmt.Sprintf("UPDATE %s SET ", *qb.table)
	for _, kv := range qb.setKv {
		q = q + kv.String() + ", "
	}

	if qb.whereKv != nil && len(qb.whereKv) > 0 {
		q = strings.TrimRight(q, ", ") + " WHERE "
	}

	for _, kv := range qb.whereKv {
		q = q + kv.String() + " AND "
	}

	if qb.returnVals != nil && len(qb.returnVals) > 0 {
		q = fmt.Sprintf("%s RETURNING %s", q, strings.Join(qb.returnVals, ", "))
	}

	return q + ";", nil
}
