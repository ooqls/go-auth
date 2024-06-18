package query

import "fmt"

type KeyValue struct {
	Key   string
	Value any
}

func (kv *KeyValue) String() string {
	return fmt.Sprintf("%s = %s", kv.Key, kv.Value)
}
