package domain

import (
	"gong/sdk"
	"gong/types"

	"github.com/oklog/ulid/v2"
)

type MsgSdkEvent struct {
	Event sdk.Event
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
}

type RspGetProjectEvents struct {
	Events types.EventList
	Props  types.ProjectActorProps
}

type MsgEventsListenStart struct {
	ID string
}

type MsgEventsListenStop struct {
	ID string
}

type MsgListenerSendBatch struct{}

type MsgSse struct {
	Msgs []any
}

type MsgSsePing struct{}

type MsgSseNewActor struct {
	ActorProps types.ProjectActorProps
}

type MsgSseNewEvents struct {
	ActorPID   string
	NumEvents  int32
	ActorAlive bool
}
