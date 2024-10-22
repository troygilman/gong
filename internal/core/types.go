package core

import (
	"context"
	"stagebase/internal/common"
	"stagebase/internal/model"
	"stagebase/internal/persistance"
	"stagebase/pkg/dolly"
	"stagebase/sdk"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	kindProject    = "project"
	kindProcessor  = "processor"
	kindListener   = "listener"
	kindController = "controller"
)

type commonActorData struct {
	id          uuid.UUID
	app         App
	poisonTimer dolly.PoisonTimer
}

const (
	ActorInitialized  = "actor.Initialized"
	ActorStarted      = "actor.Started"
	ActorStopped      = "actor.Stopped"
	ActorStoppedEvent = "actor.ActorStoppedEvent"
)

func isActorAlive(eventType string) bool {
	return eventType != ActorStopped && eventType != ActorStoppedEvent
}

type App interface {
	Shutdown()
	PID(kind string, id string) (*actor.PID, error)
	Database() persistance.Database
	JWT(tokenString string) (jwt.RegisteredClaims, error)
	Actor(ctx context.Context, projectID uuid.UUID, actorPID string) (actor model.Actor, err error)
	ProjectProps(ctx context.Context, projectID uuid.UUID) (project model.ProjectProps, err error)
	Events(ctx context.Context, projectID uuid.UUID, actorPID string, timestamp time.Time, filter model.Filter, pageIndex int) (events []model.Event, maxPageIndex int, err error)
	ProcessEvents(ctx context.Context, projectID uuid.UUID, events []sdk.Event) error
	// SourcesSinks(ctx context.Context, projectID uuid.UUID, actorPID string, sources bool, rng Range) (pids []string, err error)
	// SourcesSinksNodeInfo(ctx context.Context, projectID uuid.UUID, actorPID string, nodePID string, source bool, rng Range) (SourcesSinksNodeInfo, error)
	ListenStart(ctx context.Context, projectID uuid.UUID, id string, lType ListenerType, filter common.Filter) (<-chan []any, error)
	ListenStop(ctx context.Context, id string) error
}

type ListenerType uint

const (
	ListenerTypeProject ListenerType = iota
	ListenerTypeEvents
	ListenerTypeMetrics
)
