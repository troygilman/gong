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

type RspGetProjectActorProps struct {
	Props types.ProjectActorProps
}

type MsgGetProjectEvents struct {
	ActorPID  string
	EventID   ulid.ULID
	NumEvents int
	Filter    types.EventsFilter
}

type RspGetProjectEvents struct {
	Events types.EventList
	Props  types.ProjectActorProps
}

type MsgEventsListenStart struct{}

type MsgEventsListenStop struct{}

type MsgSendUpdates struct{}

type MsgSse struct {
	Msgs []any
}

type MsgSseNewActor struct {
	ActorProps types.ProjectActorProps
}

type MsgSseNewEvents struct {
	ActorProps types.ProjectActorProps
}

type MsgUpdateProjectActorProps struct {
	Props []types.ProjectActorProps
}
