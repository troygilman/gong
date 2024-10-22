package core

import (
	"context"
	"stagebase/internal/common"
	"stagebase/internal/persistance"
	"stagebase/pkg/dolly"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	KindProject    = "project"
	KindProcessor  = "processor"
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
	Send(pid *actor.PID, msg any)
	Request(pid *actor.PID, msg any) (any, error)
	Database() persistance.Database
	JWT(tokenString string) (jwt.RegisteredClaims, error)
	ListenStart(ctx context.Context, projectID uuid.UUID, id string, lType ListenerType, filter common.Filter) (<-chan []any, error)
	ListenStop(ctx context.Context, id string) error
}

type ListenerType uint

const (
	ListenerTypeProject ListenerType = iota
	ListenerTypeEvents
	ListenerTypeMetrics
)
