package domain

import (
	"gong/sdk"
	"gong/types"

	"github.com/oklog/ulid/v2"
)

type MsgSdkEvents struct {
	Events []sdk.Event
}

type MsgEvent struct {
	Event types.Event
}

type MsgGetProjectProps struct{}

type RspGetProjectProps struct {
	Props types.ProjectProps
}

type MsgGetProjectActorProps struct {
	ActorPID string
}

type RspGetProjectActor struct {
	Actor types.Actor
}

type MsgGetProjectEvents struct {
	ActorPID  string
	EventID   ulid.ULID
	NumEvents int
	Filter    types.EventsFilter
}

type RspGetProjectEvents struct {
	Events types.EventList
	Actor  types.Actor
}

type MsgEventsListenStart struct{}

type MsgEventsListenStop struct{}

type msgProcessorUpdateState struct{}

type MsgSse struct {
	Msgs []any
}

type MsgSseNewActor struct {
	Actor types.Actor
}

type MsgSseNewEvents struct {
	Actor types.Actor
}

type MsgUpdateProjectActors struct {
	Actors []types.Actor
}

type msgProjectUpdateState struct{}

type MsgGetActorSourcesSinks struct {
	ActorPID string
	Flow     string
}

type RspGetActorSourcesSinks struct {
	SourcesSinks []string
}

type msgListenStart struct {
	id string
	c  chan<- []any
}

type msgListenStop struct {
	id string
}
