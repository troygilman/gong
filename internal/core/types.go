package core

import (
	"stagebase/internal"
	"stagebase/pkg/dolly"

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
	app         internal.App
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
