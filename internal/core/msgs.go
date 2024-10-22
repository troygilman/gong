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

type getProjectEventsMsg struct {
	actorPID  string
	eventID   ulid.ULID
	timestamp time.Time
	pageIndex int
	filter    model.Filter
}

type getProjectEventsRsp struct {
	events       []model.Event
	maxPageIndex int
	actor        model.Actor
}

type processorUpdateStateMsg struct{}

type updateProjectMsg struct {
	actors []model.Actor
	events []model.Event
	metric Metric
}

type getActorSourcesSinksMsg struct {
	actorPID string
	sources  bool
	rng      common.Range
}

type getActorSourcesSinksRsp struct {
	sourcesSinks []string
}

type getSourcesSinksNodeInfoMsg struct {
	actorPID string
	nodePID  string
	source   bool
	rng      common.Range
}

type getSourcesSinksNodeInfoRsp struct {
	info model.SourcesSinksNodeInfo
}

type listenStartMsg struct {
	lType ListenerType
}

type listenStopMsg struct {
	lType ListenerType
}

type listenDataMsg struct {
	data []any
}

type SseNewActorMsg struct {
	Actor model.Actor
}

type SseNewEventsMsg struct {
	Actor model.Actor
}
