package pg

import (
	"reflect"
)

func GetAllDbFields[T any](obj T) []string {
	var fields []string
	oType := reflect.TypeOf(obj)
	for i := 0; i < oType.NumField(); i++ {
		f := oType.Field(i)
		v := f.Tag.Get("db")
		if v != "" {
			fields = append(fields, v)
		}
	}

	return fields
}
