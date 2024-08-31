package common

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type CtxKey string

const (
	CtxKeyApp       CtxKey = "app"
	CtxKeyProjectID CtxKey = "projectID"
	CtxKeyPID       CtxKey = ParamPID
	CtxKeyTab       CtxKey = ParamTab
	CtxKeyFilter    CtxKey = "filter"
	CtxKeyFlow      CtxKey = ParamFlow
	CtxKeyFlowNode  CtxKey = ParamFlowNode
	CtxKeyFlowRange CtxKey = ParamFlowRange
)

func valueHelper[T any](ctx context.Context, key CtxKey) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}

func mustValueHelper[T any](ctx context.Context, key CtxKey) (val T) {
	if val, ok := valueHelper[T](ctx, key); ok {
		return val
	} else {
		panic(fmt.Sprintf("value %s of type %T does not exist in context", key, val))
	}
}

func ValueApp(ctx context.Context) (App, bool) {
	return valueHelper[App](ctx, CtxKeyApp)
}

func MustValueApp(ctx context.Context) App {
	return mustValueHelper[App](ctx, CtxKeyApp)
}

func ValueProjectID(ctx context.Context) (uuid.UUID, bool) {
	return valueHelper[uuid.UUID](ctx, CtxKeyProjectID)
}

func MustValueProjectID(ctx context.Context) uuid.UUID {
	return mustValueHelper[uuid.UUID](ctx, CtxKeyProjectID)
}

func MustValueActorPID(ctx context.Context) string {
	return mustValueHelper[string](ctx, CtxKeyPID)
}
