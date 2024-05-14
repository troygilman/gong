package types

import (
	"context"
	"gong/sdk"
	"net/url"
	"sort"

	"github.com/google/uuid"
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
	ID         uuid.UUID
	ActorProps map[string]ProjectActorProps
}

type ProjectActorProps struct {
	PID          string
	NumEvents    int32
	Alive        bool
	FirstEventID ulid.ULID
	LastEventID  ulid.ULID
}

type EventsFilter struct {
	Type   string
	Sender string
}

func (f EventsFilter) SetValues(values url.Values) {
	if f.Type != "" {
		values.Set("filter_type", f.Type)
	}
	if f.Sender != "" {
		values.Set("filter_sender", f.Sender)
	}
}
