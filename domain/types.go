package domain

import (
	"gong/sdk"

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

type EventHeap []sdk.Event

func (h EventHeap) Len() int           { return len(h) }
func (h EventHeap) Less(i, j int) bool { return h[j].Timestamp.After(h[i].Timestamp) }
func (h EventHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *EventHeap) Push(x any) {
	*h = append(*h, x.(sdk.Event))
}

func (h *EventHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type ProjectProps struct {
	ID     uuid.UUID
	Actors []string
}
