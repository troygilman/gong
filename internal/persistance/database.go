package persistance

import (
	"context"
	"stagebase/internal/model"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -source=database.go -destination=database_mock.go -package=persistance

type Database interface {
	QueryEvents(ctx context.Context, query EventsQuery) ([]model.Event, int, error)
	CreateEvents(ctx context.Context, events []model.Event) error
	QueryActors(ctx context.Context, projectID uuid.UUID) ([]model.Actor, error)
	CreateActor(ctx context.Context, actor model.Actor) error
	UpdateActor(ctx context.Context, actor model.Actor) error
	QueryGraphLinks(ctx context.Context, query GraphLinksQuery) ([]model.GraphLink, error)
	QuerySourcesSinksNodeInfo(context.Context, SourcesSinksNodeInfoQuery) (model.SourcesSinksNodeInfo, error)
}

type EventsQuery struct {
	ProjectID uuid.UUID
	From      time.Time
	To        time.Time
	PageIndex int
	Filter    model.Filter
}

type GraphLinksQuery struct {
	ProjectID uuid.UUID
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
