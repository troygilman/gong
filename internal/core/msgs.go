package core

import (
	"stagebase/internal"
	"stagebase/sdk"
	"time"

	"github.com/oklog/ulid/v2"
)

type msgSdkEvents struct {
	events []sdk.Event
}

type msgGetProjectProps struct{}

type rspGetProjectProps struct {
	props internal.ProjectProps
}

type msgGetProjectActorProps struct {
	actorPID string
}

type rspGetProjectActor struct {
	actor internal.Actor
}

type msgGetProjectEvents struct {
	actorPID  string
	eventID   ulid.ULID
	timestamp time.Time
	pageIndex int
	filter    internal.Filter
}

type rspGetProjectEvents struct {
	events       internal.EventList
	maxPageIndex int
	actor        internal.Actor
}

type msgProcessorUpdateState struct{}

type msgUpdateProject struct {
	actors []internal.Actor
	events []internal.Event
	metric Metric
}

type msgGetActorSourcesSinks struct {
	actorPID string
	sources  bool
	rng      internal.Range
}

type rspGetActorSourcesSinks struct {
	sourcesSinks []string
}

type msgGetSourcesSinksNodeInfo struct {
	actorPID string
	nodePID  string
	source   bool
	rng      internal.Range
}

type rspGetSourcesSinksNodeInfo struct {
	info internal.SourcesSinksNodeInfo
}

type msgListenStart struct {
	lType internal.ListenerType
}

type msgListenStop struct {
	lType internal.ListenerType
}

type msgListenData struct {
	data []any
}
