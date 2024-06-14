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
	ParamPID          = "pid"
	ParamTab          = "tab"
	ParamFilterID     = "filter_id"
	ParamFilterType   = "filter_type"
	ParamFilterSender = "filter_sender"
	ParamFilterLevel  = "filter_level"
	ParamFlow         = "flow"
	ParamFlowNode     = "flow_node"
)

type CtxKey string

const (
	CtxProject  CtxKey = "projectID"
	CtxPID      CtxKey = ParamPID
	CtxTab      CtxKey = ParamTab
	CtxFilter   CtxKey = "filter"
	CtxFlow     CtxKey = ParamFlow
	CtxFlowNode CtxKey = ParamFlowNode
)

const (
	ProjectViewActor             = "actor"
	ProjectViewFlowChart         = "flow-chart"
	ProjectViewFlowChartNodeInfo = "flow-chart-node-info"
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
	ID     *string
	Type   *string
	Sender *string
	Level  *slog.Level
}

func NewFilter(ctx echo.Context) (filter Filter) {
	id := ctx.QueryParam(ParamFilterID)
	if id != "" {
		filter.ID = &id
	}

	fType := ctx.QueryParam(ParamFilterType)
	if fType != "" {
		filter.Type = &fType
	}

	sender := ctx.QueryParam(ParamFilterSender)
	if sender != "" {
		filter.Sender = &sender
	}

	level, err := strconv.Atoi(ctx.QueryParam(ParamFilterLevel))
	if err == nil {
		sLevel := slog.Level(level)
		filter.Level = &sLevel
	}

	return filter
}

func (filter Filter) ValidEvent(event Event) bool {
	if filter.ID != nil && *filter.ID != event.ID.String() {
		return false
	}

	if filter.Type != nil && *filter.Type != event.Type {
		return false
	}

	if filter.Sender != nil && (event.Sender == nil || *filter.Sender != *event.Sender) {
		return false
	}

	if filter.Level != nil && *filter.Level > event.Level {
		return false
	}

	return true
}
