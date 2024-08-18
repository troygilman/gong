package persistance

import (
	"context"
	"gong/internal/common"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

//go:generate mockgen -source=database.go -destination=database_mock.go -package=persistance

type Database interface {
	QueryEvents(ctx context.Context, query EventsQuery) ([]common.Event, error)
	CreateEvents(ctx context.Context, events []common.Event) error
	QueryActors(ctx context.Context, projectID uuid.UUID) ([]common.Actor, error)
	CreateActor(ctx context.Context, actor common.Actor) error
	UpdateActor(ctx context.Context, actor common.Actor) error
	QueryActorSourcesSinks(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error)
	QuerySourcesSinksNodeInfo(context.Context, SourcesSinksNodeInfoQuery) (common.SourcesSinksNodeInfo, error)
}

type EventsQuery struct {
	ProjectID uuid.UUID
	ActorPID  string
	EventID   ulid.ULID
	NumEvents int
	Timestamp time.Time
	Filter    common.Filter
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
