package domain

import (
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

type persistentActor interface {
	load(id uuid.UUID) error
	create(id uuid.UUID) error
}

type commonActorData struct {
	id          uuid.UUID
	app         App
	poisonTimer actor_utils.PoisonTimer
}
