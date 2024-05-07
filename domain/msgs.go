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
}

type RspGetProjectEvents struct {
	Events types.EventList
}

type MsgEventsListenStart struct {
	ID  string
	SSE chan<- string
}

type MsgEventsListenStop struct {
	ID string
}

type MsgListenerSendBatch struct{}
