package domain

import (
	"gong/sdk"
	"gong/types"
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
	ActorPID string
	StartIdx int32
	EndIdx   int32
}

type RspGetProjectEvents struct {
	Events         types.EventList
	NumTotalEvents int32
}

type MsgEventsListenStart struct {
	ID  string
	SSE chan<- any
}

type MsgEventsListenStop struct {
	ID string
}

type MsgListenerSendBatch struct{}

type MsgSsePing struct{}

type MsgSseNewActor struct {
	ActorProps types.ProjectActorProps
}

type MsgSseNewEvents struct {
	ActorPID   string
	NumEvents  int32
	ActorAlive bool
}
