package persistance

import (
	"context"
	"gong/types"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type Database interface {
	QueryEvents(ctx context.Context, query EventsQuery) ([]types.Event, error)
	CreateEvents(ctx context.Context, events []types.Event) error
	QueryActors(ctx context.Context, projectID uuid.UUID) ([]types.Actor, error)
	CreateActor(ctx context.Context, actor types.Actor) error
	UpdateActor(ctx context.Context, actor types.Actor) error
	QueryActorSources(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error)
	QueryActorSinks(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error)
}

type EventsQuery struct {
	ProjectID uuid.UUID
	ActorPID  string
	EventID   ulid.ULID
	NumEvents int
	Filter    types.EventsFilter
}

type ActorSourcesSinksQuery struct {
	ProjectID uuid.UUID
	ActorPID  string
	MinTime   time.Time
	MaxTime   time.Time
}
