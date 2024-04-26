package domain

import "gong/sdk"

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
	Events EventList
}

type MsgEventsListenStart struct {
	ID     string
	Events chan<- sdk.Event
}

type MsgEventsListenStop struct {
	ID string
}
