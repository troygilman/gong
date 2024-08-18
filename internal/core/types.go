package core

import (
	"context"
	"gong/internal/common"
	"gong/internal/persistance"
	"gong/pkg/dolly"
	"gong/sdk"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type App interface {
	Shutdown() error
	PID(kind string, id string) (*actor.PID, error)
	Database() persistance.Database
	JWT(tokenString string) (jwt.RegisteredClaims, error)
	Actor(ctx context.Context, projectID uuid.UUID, actorPID string) (actor common.Actor, err error)
	ProjectProps(ctx context.Context, projectID uuid.UUID) (project common.ProjectProps, err error)
	Events(ctx context.Context, projectID uuid.UUID, actorPID string, timestamp time.Time, numEvents int, filter common.Filter) (events common.EventList, err error)
	ProcessEvents(ctx context.Context, projectID uuid.UUID, events []sdk.Event) error
	SourcesSinks(ctx context.Context, projectID uuid.UUID, actorPID string, sources bool, rng common.Range) (pids []string, err error)
	SourcesSinksNodeInfo(ctx context.Context, projectID uuid.UUID, actorPID string, nodePID string, source bool, rng common.Range) (common.SourcesSinksNodeInfo, error)
	ListenStart(ctx context.Context, projectID uuid.UUID, id string, lType ListenerType, filter common.Filter) (<-chan []any, error)
	ListenStop(ctx context.Context, id string) error
}

const (
	kindProject    = "project"
	kindProcessor  = "processor"
	kindListener   = "listener"
	kindController = "controller"
)

type ListenerType uint

const (
	ListenerTypeProject ListenerType = iota
	ListenerTypeEvents
	ListenerTypeMetrics
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
