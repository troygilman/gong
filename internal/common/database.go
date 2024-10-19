package common

import (
	"context"
	"stagebase/internal/core/model"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

//go:generate mockgen -source=database.go -destination=database_mock.go -package=common

type Database interface {
	QueryEvents(ctx context.Context, query EventsQuery) ([]Event, int, error)
	CreateEvents(ctx context.Context, events []Event) error
	QueryActors(ctx context.Context, projectID uuid.UUID) ([]model.Actor, error)
	CreateActor(ctx context.Context, actor model.Actor) error
	UpdateActor(ctx context.Context, actor model.Actor) error
	QueryActorSourcesSinks(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error)
	QuerySourcesSinksNodeInfo(context.Context, SourcesSinksNodeInfoQuery) (SourcesSinksNodeInfo, error)
}

type EventsQuery struct {
	ProjectID uuid.UUID
	ActorPID  string
	EventID   ulid.ULID
	Timestamp time.Time
	PageIndex int
	Filter    model.Filter
}

type ActorSourcesSinksQuery struct {
	Sources   bool
	ProjectID uuid.UUID
	ActorPID  string
	MinTime   time.Time
	MaxTime   time.Time
}

type SourcesSinksNodeInfoQuery struct {
	Source    bool
	ProjectID uuid.UUID
	ActorPID  string
	NodePID   string
	MinTime   time.Time
	MaxTime   time.Time
}
