package common

import (
	"gong/common"
	"gong/sdk"
	"time"

	"github.com/oklog/ulid/v2"
)

type MsgSdkEvents struct {
	Events []sdk.Event
}

type MsgGetProjectProps struct{}

type RspGetProjectProps struct {
	Props common.ProjectProps
}

type MsgGetProjectActorProps struct {
	ActorPID string
}

type RspGetProjectActor struct {
	Actor common.Actor
}

type MsgGetProjectEvents struct {
	ActorPID  string
	EventID   ulid.ULID
	Timestamp time.Time
	NumEvents int
	Filter    common.Filter
}

type RspGetProjectEvents struct {
	Events common.EventList
	Actor  common.Actor
}

type MsgProcessorUpdateState struct{}

type MsgUpdateProject struct {
	Actors []common.Actor
	Events []common.Event
	Metric Metric
}

type MsgGetActorSourcesSinks struct {
	ActorPID string
	Sources  bool
}

type RspGetActorSourcesSinks struct {
	SourcesSinks []string
}

type MsgGetSourcesSinksNodeInfo struct {
	ActorPID string
	NodePID  string
	Source   bool
}

type RspGetSourcesSinksNodeInfo struct {
	Info common.SourcesSinksNodeInfo
}

type MsgListenStart struct {
	LType ListenerType
}

type MsgListenStop struct {
	LType ListenerType
}

type MsgListenData struct {
	Data []any
}
