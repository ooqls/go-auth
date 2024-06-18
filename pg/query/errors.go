package query

import "errors"

var MissingTableErr error = errors.New("table not given")
var MissingColumnsErr error = errors.New("columns not given")
var MissingSetValuesErr error = errors.New("set values not given")
var MissingWhereClauseErr error = errors.New("where clause not given and all flag not set")
