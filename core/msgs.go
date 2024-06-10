package core

import (
	"gong/common"
	"gong/sdk"

	"github.com/oklog/ulid/v2"
)

type msgSdkEvents struct {
	events []sdk.Event
}

type msgGetProjectProps struct{}

type rspGetProjectProps struct {
	props common.ProjectProps
}

type msgGetProjectActorProps struct {
	actorPID string
}

type rspGetProjectActor struct {
	actor common.Actor
}

type msgGetProjectEvents struct {
	actorPID  string
	eventID   ulid.ULID
	numEvents int
	filter    common.Filter
}

type rspGetProjectEvents struct {
	events common.EventList
	actor  common.Actor
}

type msgProcessorUpdateState struct{}

type msgUpdateProject struct {
	actors []common.Actor
	events []common.Event
}

type msgUpdateProjectInternal struct{}

type msgGetActorSourcesSinks struct {
	actorPID string
	sources  bool
}

type rspGetActorSourcesSinks struct {
	sourcesSinks []string
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
