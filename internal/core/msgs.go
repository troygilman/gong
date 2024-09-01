package core

import (
	"gong/internal/common"
	"gong/sdk"
	"time"

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
	timestamp time.Time
	pageIndex int
	filter    common.Filter
}

type rspGetProjectEvents struct {
	events       common.EventList
	maxPageIndex int
	actor        common.Actor
}

type msgProcessorUpdateState struct{}

type msgUpdateProject struct {
	actors []common.Actor
	events []common.Event
	metric Metric
}

type msgGetActorSourcesSinks struct {
	actorPID string
	sources  bool
	rng      common.Range
}

type rspGetActorSourcesSinks struct {
	sourcesSinks []string
}

type msgGetSourcesSinksNodeInfo struct {
	actorPID string
	nodePID  string
	source   bool
	rng      common.Range
}

type rspGetSourcesSinksNodeInfo struct {
	info common.SourcesSinksNodeInfo
}

type msgListenStart struct {
	lType common.ListenerType
}

type msgListenStop struct {
	lType common.ListenerType
}

type msgListenData struct {
	data []any
}
