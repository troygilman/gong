package core

import (
	"stagebase/internal/common"
	"stagebase/internal/model"
	"stagebase/sdk"
	"time"

	"github.com/oklog/ulid/v2"
)

type ProcessEventsMsg struct {
	Events []sdk.Event
}

type GetProjectPropsMsg struct{}

type GetActorMsg struct {
	Pid string
}

type msgGetProjectEvents struct {
	actorPID  string
	eventID   ulid.ULID
	timestamp time.Time
	pageIndex int
	filter    model.Filter
}

type rspGetProjectEvents struct {
	events       []model.Event
	maxPageIndex int
	actor        model.Actor
}

type msgProcessorUpdateState struct{}

type msgUpdateProject struct {
	actors []model.Actor
	events []model.Event
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
	info model.SourcesSinksNodeInfo
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

type SseNewActorMsg struct {
	Actor model.Actor
}

type SseNewEventsMsg struct {
	Actor model.Actor
}
