package common

import (
	"context"
	"errors"
	"gong/sdk"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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
	ParamFlowRange    = "flow_range"
)

type CtxKey string

const (
	CtxProject   CtxKey = "projectID"
	CtxPID       CtxKey = ParamPID
	CtxTab       CtxKey = ParamTab
	CtxFilter    CtxKey = "filter"
	CtxFlow      CtxKey = ParamFlow
	CtxFlowNode  CtxKey = ParamFlowNode
	CtxFlowRange CtxKey = ParamFlowRange
)

const (
	ProjectViewActor             = "actor"
	ProjectViewFlowChart         = "flow-chart"
	ProjectViewFlowChartNodeInfo = "flow-chart-node-info"
)

type Range int

func (r Range) String() string {
	return strconv.Itoa(int(r))
}

func (r *Range) Parse(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*r = Range(i)
	return nil
}

func (r Range) Duration() (time.Duration, error) {
	switch r {
	case 0:
		return time.Minute, nil
	case 1:
		return time.Hour, nil
	case 2:
		return time.Hour * 24, nil
	case 3:
		return time.Hour * 24 * 7, nil
	case 4:
		return time.Hour * 24 * 30, nil
	default:
		return 0, errors.New("invalid range")
	}
}

const (
	RangeMinute Range = iota
	RangeHour
	RangeDay
	RangeWeek
	RangeMonth
)

type Event struct {
	ProjectID string
	sdk.Event
}

type EventList []Event

func (l *EventList) Push(e Event) {
	i := sort.Search(len(*l), func(i int) bool {
		return e.Timestamp.Before((*l)[i].Timestamp)
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
	NumEvents   int
	Alive       bool
	LastEventID string
}

type Filter struct {
	PID    *string
	Type   *string
	Sender *string
	Level  *slog.Level
}

func NewFilter(ctx echo.Context) (filter Filter) {
	pid := ctx.QueryParam(ParamPID)
	if pid != "" {
		filter.PID = &pid
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
	if filter.PID != nil && *filter.PID != event.PID {
		return false
	}

	if filter.Type != nil && *filter.Type != event.Type {
		return false
	}

	if filter.Sender != nil && *filter.Sender != event.Sender {
		return false
	}

	if filter.Level != nil && *filter.Level > event.Level {
		return false
	}

	return true
}

type SourcesSinksNodeInfo struct {
	PID      string
	MsgCount int64
}
