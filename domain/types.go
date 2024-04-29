package domain

import (
	"gong/sdk"
	"sort"

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

type EventList []sdk.Event

func (l *EventList) Push(e sdk.Event) {
	i := sort.Search(len(*l), func(i int) bool {
		return e.Timestamp.Before((*l)[i].Timestamp)
	})
	*l = append(*l, sdk.Event{})
	copy((*l)[i+1:], (*l)[i:])
	(*l)[i] = e
}

type ProjectProps struct {
	ID         uuid.UUID
	ActorProps []ProjectActorProps
}

type ProjectActorProps struct {
	PID       string
	NumEvents int
}

type ListenerEvent struct {
	Event     sdk.Event
	NumEvents int
}
