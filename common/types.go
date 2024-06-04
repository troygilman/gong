package common

import (
	"context"
	"gong/sdk"
	"log/slog"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

const (
	SseTypeNewActor             = "new-actor"
	SseTypeNewEvents            = "new-events:"
	SseTypeDeleteActor          = "delete-actor:"
	SseTypeUpdateActorNumEvents = "update-actor-num-events:"
)

type CtxKey string

const (
	CtxProject CtxKey = "project"
	CtxPID     CtxKey = "pid"
	CtxTab     CtxKey = "tab"
	CtxFilter  CtxKey = "filter"
	CtxFlow    CtxKey = "flow"
)

const (
	ProjectViewActor     = "actor"
	ProjectViewFlowChart = "flow-chart"
)

type Event struct {
	ID        ulid.ULID
	ProjectID uuid.UUID
	Index     int32
	sdk.Event
}

type EventList []Event

func (l *EventList) Push(e Event) {
	i := sort.Search(len(*l), func(i int) bool {
		return e.Index < (*l)[i].Index
		//return e.Timestamp.Before((*l)[i].Timestamp)
	})
	*l = append(*l, Event{})
	copy((*l)[i+1:], (*l)[i:])
	(*l)[i] = e
}

type ContextBuilder struct {
	ctx context.Context
}

func NewContextBuilder(ctx context.Context) ContextBuilder {
	return ContextBuilder{
		ctx: ctx,
	}
}

func (cb ContextBuilder) With(key CtxKey, val any) ContextBuilder {
	cb.ctx = context.WithValue(cb.ctx, key, val)
	return cb
}

func (cb ContextBuilder) Build() context.Context {
	return cb.ctx
}

type ProjectProps struct {
	ID     uuid.UUID
	Actors map[string]Actor
}

type Actor struct {
	PID         string
	ProjectID   uuid.UUID
	NumEvents   int32
	Alive       bool
	LastEventID ulid.ULID
}

type Filter struct {
	ID     string
	Type   string
	Sender string
	Level  slog.Level
}

func NewFilter(ctx echo.Context) Filter {
	level, err := strconv.Atoi(ctx.QueryParam("filter_level"))
	if err != nil {
		level = int(slog.LevelDebug)
	}

	return Filter{
		ID:     ctx.QueryParam("filter_id"),
		Type:   ctx.QueryParam("filter_type"),
		Sender: ctx.QueryParam("filter_sender"),
		Level:  slog.Level(level),
	}
}
