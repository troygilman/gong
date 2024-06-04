package core

import (
	"context"

	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	kindProject   = "project"
	kindProcessor = "processor"
	kindListener  = "listener"
)

type persistentActor interface {
	load(ctx context.Context, id uuid.UUID) error
	create(ctx context.Context, id uuid.UUID) error
}

type commonActorData struct {
	id          uuid.UUID
	app         App
	poisonTimer actor_utils.PoisonTimer
}
