package insert

import (
	"fmt"
	"strings"

	"github.com/braumsmilk/go-auth/pg/query"
)

func BuildInsertQuery(table string, cols ...string) string {
	sqlCols := strings.Join(cols, ", ")
	namedCol := []string{}
	for _, c := range cols {
		namedCol = append(namedCol, ":"+c)
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", table, sqlCols, strings.Join(namedCol, ", "))
}

type QueryBuilder struct {
	table   *string
	columns []string
	values  []any
}

func (qb *QueryBuilder) Insert(table string, columns []string) *QueryBuilder {
	qb.table = &table
	qb.columns = columns
	return qb
}

func (qb *QueryBuilder) Build() (string, error) {
	if qb.table == nil {
		return "", query.MissingTableErr
	}

	if qb.columns == nil || len(qb.columns) == 0 {
		return "", query.MissingColumnsErr
	}

	values := []string{}
	for _, c := range qb.columns {
		values = append(values, ":"+c)
	}

	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s) ",
		*qb.table,
		strings.Join(qb.columns, ", "),
		strings.Join(values, ","))

	return q, nil
}
