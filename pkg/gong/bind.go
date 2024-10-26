package gong

import (
	"context"
	"fmt"
)

func Bind(ctx context.Context, binds ...BindFunc) error {
	for _, bind := range binds {
		if err := bind(ctx); err != nil {
			return err
		}
	}
	return nil
}

type BindFunc func(context.Context) error

func NewBinding[T any](key string, ref *T) BindFunc {
	return func(ctx context.Context) error {
		val, ok := ctx.Value(key).(T)
		if !ok {
			return fmt.Errorf("value %s of type %T does not exist in context", key, val)
		}
		*ref = val
		return nil
	}
}
