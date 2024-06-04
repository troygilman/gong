package core

import (
	"gong/common"
	"gong/sdk"

	"github.com/oklog/ulid/v2"
)

type msgSdkEvents struct {
	Events []sdk.Event
}

type msgGetProjectProps struct{}

type rspGetProjectProps struct {
	Props common.ProjectProps
}

type msgGetProjectActorProps struct {
	ActorPID string
}

type rspGetProjectActor struct {
	Actor common.Actor
}

type msgGetProjectEvents struct {
	ActorPID  string
	EventID   ulid.ULID
	NumEvents int
	Filter    common.Filter
}

type rspGetProjectEvents struct {
	Events common.EventList
	Actor  common.Actor
}

type msgProcessorUpdateState struct{}

type msgUpdateProjectActors struct {
	Actors []common.Actor
}

type msgProjectUpdateState struct{}

type msgGetActorSourcesSinks struct {
	ActorPID string
	Sources  bool
}

type rspGetActorSourcesSinks struct {
	SourcesSinks []string
}

type msgListenStart struct {
	lType ListenerType
}

type msgListenStop struct {
	lType ListenerType
}

type msgListenData struct {
	data []any
}
