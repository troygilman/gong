package domain

import (
	"gong/sdk"
	"gong/types"
)

type MsgEvent struct {
	Event sdk.Event
}

type MsgGetProjectProps struct{}

type RspGetProjectProps struct {
	Props ProjectProps
}

type MsgGetProjectEvents struct {
	ActorPID string
}

type RspGetProjectEvents struct {
	Events types.EventList
}

type MsgEventsListenStart struct {
	ID     string
	Events chan<- types.Event
}

type MsgEventsListenStop struct {
	ID string
}
