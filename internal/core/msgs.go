package core

import (
	"stagebase/internal/model"
	"stagebase/sdk"
)

type ProcessEventsMsg struct {
	Events []sdk.Event
}

type GetProjectPropsMsg struct{}

type GetActorMsg struct {
	Pid string
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
