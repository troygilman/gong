package common

import (
	"context"
	"gong/common"
	"gong/persistance"
	"gong/sdk"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

type App interface {
	PID(kind string, id string) (*actor.PID, error)
	Database() persistance.Database
	JWT(tokenString string) (jwt.RegisteredClaims, error)
	Actor(ctx context.Context, projectID uuid.UUID, actorPID string) (actor common.Actor, err error)
	ProjectProps(ctx context.Context, projectID uuid.UUID) (project common.ProjectProps, err error)
	Events(ctx context.Context, projectID uuid.UUID, actorPID string, timestamp time.Time, numEvents int, filter common.Filter) (events common.EventList, err error)
	ProcessEvents(ctx context.Context, projectID uuid.UUID, events []sdk.Event) error
	SourcesSinks(ctx context.Context, projectID uuid.UUID, actorPID string, sources bool) (pids []string, err error)
	SourcesSinksNodeInfo(ctx context.Context, projectID uuid.UUID, actorPID string, nodePID string, source bool) (common.SourcesSinksNodeInfo, error)
	ListenStart(ctx context.Context, projectID uuid.UUID, id string, lType ListenerType, filter common.Filter) (<-chan []any, error)
	ListenStop(ctx context.Context, id string) error
}

const (
	KindProject   = "project"
	KindProcessor = "processor"
	KindListener  = "listener"
)

type ListenerType uint

const (
	ListenerTypeProject ListenerType = iota
	ListenerTypeEvents
	ListenerTypeMetrics
)

type PersistentActor interface {
	Load(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, id uuid.UUID) error
}

type CommonActorData struct {
	Id          uuid.UUID
	App         App
	PoisonTimer actor_utils.PoisonTimer
}

const (
	ActorInitialized  = "actor.Initialized"
	ActorStarted      = "actor.Started"
	ActorStopped      = "actor.Stopped"
	ActorStoppedEvent = "actor.ActorStoppedEvent"
)

func IsActorAlive(eventType string) bool {
	return eventType != ActorStopped && eventType != ActorStoppedEvent
}
