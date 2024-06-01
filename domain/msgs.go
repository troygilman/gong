package domain

import (
	"gong/sdk"
	"gong/types"

	"github.com/oklog/ulid/v2"
)

type msgSdkEvents struct {
	Events []sdk.Event
}

type msgGetProjectProps struct{}

type rspGetProjectProps struct {
	Props types.ProjectProps
}

type msgGetProjectActorProps struct {
	ActorPID string
}

type rspGetProjectActor struct {
	Actor types.Actor
}

type msgGetProjectEvents struct {
	ActorPID  string
	EventID   ulid.ULID
	NumEvents int
	Filter    types.EventsFilter
}

type rspGetProjectEvents struct {
	Events types.EventList
	Actor  types.Actor
}

type msgProcessorUpdateState struct{}

type msgUpdateProjectActors struct {
	Actors []types.Actor
}

type msgProjectUpdateState struct{}

type msgGetActorSourcesSinks struct {
	ActorPID string
	Flow     string
}

type rspGetActorSourcesSinks struct {
	SourcesSinks []string
}

type msgListenStart struct {
	id string
	c  chan<- []any
}

type msgListenStop struct {
	id string
}
