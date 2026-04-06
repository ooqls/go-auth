package contexts

import "context"
	
type Context[T any] struct {
	context.Context
	userId string
	Value  T
}

func NewContext[T any](ctx context.Context, userId string, value T) *Context[T] {
	return &Context[T]{
		Context: ctx,
		userId:  userId,
		Value:   value,
	}
}
