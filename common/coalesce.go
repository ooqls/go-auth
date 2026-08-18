package common

func Coalesce[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}

	return nil
}
