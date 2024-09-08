package common

import (
	"context"
	"fmt"
	"time"

	"stagebase/pkg/gong"

	"github.com/google/uuid"
)

const (
	CtxKeyRequestID gong.Key = "request_id"
	CtxKeyLogger    gong.Key = "logger"
	CtxKeyApp       gong.Key = "app"
	CtxKeyProjectID gong.Key = "projectID"
	CtxKeyPID       gong.Key = ParamPID
	CtxKeyTab       gong.Key = ParamTab
	CtxKeyFilter    gong.Key = "filter"
	CtxKeyFlow      gong.Key = ParamFlow
	CtxKeyFlowNode  gong.Key = ParamFlowNode
	CtxKeyFlowRange gong.Key = ParamFlowRange
	CtxKeyTimeStart gong.Key = ParamTimeStart
	CtxKeyPageIndex gong.Key = ParamPageIndex
)

func valueHelper[T any](ctx context.Context, key gong.Key) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}

func mustValueHelper[T any](ctx context.Context, key gong.Key) (val T) {
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

func CtxPID(ctx context.Context) (string, bool) {
	return valueHelper[string](ctx, CtxKeyPID)
}

func MustValueActorPID(ctx context.Context) string {
	return mustValueHelper[string](ctx, CtxKeyPID)
}

func ValueFilter(ctx context.Context) (Filter, bool) {
	return valueHelper[Filter](ctx, CtxKeyFilter)
}

func ValueTimeStart(ctx context.Context) (time.Time, bool) {
	return valueHelper[time.Time](ctx, CtxKeyTimeStart)
}

func MustValueTimeStart(ctx context.Context) time.Time {
	return mustValueHelper[time.Time](ctx, CtxKeyTimeStart)
}

func ValuePageIndex(ctx context.Context) (int, bool) {
	return valueHelper[int](ctx, CtxKeyPageIndex)
}
