package common

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CtxKey string

func (k CtxKey) String() string {
	return string(k)
}

const (
	CtxKeyRequestID CtxKey = "request_id"
	CtxKeyLogger    CtxKey = "logger"
	CtxKeyApp       CtxKey = "app"
	CtxKeyProjectID CtxKey = "projectID"
	CtxKeyPID       CtxKey = ParamPID
	CtxKeyTab       CtxKey = ParamTab
	CtxKeyFilter    CtxKey = "filter"
	CtxKeyFlow      CtxKey = ParamFlow
	CtxKeyFlowNode  CtxKey = ParamFlowNode
	CtxKeyFlowRange CtxKey = ParamFlowRange
	CtxKeyTimeStart CtxKey = ParamTimeStart
	CtxKeyPageIndex CtxKey = ParamPageIndex
)

func ctxHelper[T any](ctx context.Context, key CtxKey) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}

func mustCtxHelper[T any](ctx context.Context, key CtxKey) (val T) {
	if val, ok := ctxHelper[T](ctx, key); ok {
		return val
	} else {
		panic(fmt.Sprintf("value %s of type %T does not exist in context", key, val))
	}
}

func CtxApp(ctx context.Context) (App, bool) {
	return ctxHelper[App](ctx, CtxKeyApp)
}

func MustCtxApp(ctx context.Context) App {
	return mustCtxHelper[App](ctx, CtxKeyApp)
}

func CtxProjectID(ctx context.Context) (uuid.UUID, bool) {
	return ctxHelper[uuid.UUID](ctx, CtxKeyProjectID)
}

func MustCtxProjectID(ctx context.Context) uuid.UUID {
	return mustCtxHelper[uuid.UUID](ctx, CtxKeyProjectID)
}

func CtxPID(ctx context.Context) (string, bool) {
	return ctxHelper[string](ctx, CtxKeyPID)
}

func MustCtxPID(ctx context.Context) string {
	return mustCtxHelper[string](ctx, CtxKeyPID)
}

func MustCtxActorPID(ctx context.Context) string {
	return mustCtxHelper[string](ctx, CtxKeyPID)
}

func CtxFilter(ctx context.Context) (Filter, bool) {
	return ctxHelper[Filter](ctx, CtxKeyFilter)
}

func CtxTimeStart(ctx context.Context) (time.Time, bool) {
	return ctxHelper[time.Time](ctx, CtxKeyTimeStart)
}

func CtxValueTimeStart(ctx context.Context) time.Time {
	return mustCtxHelper[time.Time](ctx, CtxKeyTimeStart)
}

func CtxPageIndex(ctx context.Context) (int, bool) {
	return ctxHelper[int](ctx, CtxKeyPageIndex)
}
