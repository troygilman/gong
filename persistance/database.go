package persistance

import (
	"context"
	"gong/common"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type Database interface {
	QueryEvents(ctx context.Context, query EventsQuery) ([]common.Event, error)
	CreateEvents(ctx context.Context, events []common.Event) error
	QueryActors(ctx context.Context, projectID uuid.UUID) ([]common.Actor, error)
	CreateActor(ctx context.Context, actor common.Actor) error
	UpdateActor(ctx context.Context, actor common.Actor) error
	QueryActorSources(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error)
	QueryActorSinks(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error)
}

type EventsQuery struct {
	ProjectID uuid.UUID
	ActorPID  string
	EventID   ulid.ULID
	NumEvents int
	Filter    common.Filter
}

type ActorSourcesSinksQuery struct {
	ProjectID uuid.UUID
	ActorPID  string
	MinTime   time.Time
	MaxTime   time.Time
}
