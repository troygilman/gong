package common

import (
	"context"
	"github.com/google/uuid"
)

func MustApp(ctx context.Context) App {
	return ctx.Value(CtxApp).(App)
}

func MustProjectID(ctx context.Context) uuid.UUID {
	return ctx.Value(CtxProject).(uuid.UUID)
}

func MustActorPID(ctx context.Context) string {
	return ctx.Value(CtxPID).(string)
}
