package pg

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func ParseRows[T any](rows *sqlx.Rows) ([]T, error) {
	var objs []T
	for rows.Next() {
		var obj T
		err := rows.StructScan(&obj)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row into struct: %v", err)
		}

		objs = append(objs, obj)
	}

	return objs, nil
}
