package types

import (
	"gong/sdk"
	"sort"

	"github.com/google/uuid"
)

const (
	SseTypeNewActor             = "new-actor"
	SseTypeNewEvents            = "new-events:"
	SseTypeDeleteActor          = "delete-actor:"
	SseTypeUpdateActorNumEvents = "update-actor-num-events:"
)

type Event struct {
	ProjectID uuid.UUID
	Session   int32
	Index     int32
	sdk.Event
}

type EventList []Event

func (l *EventList) Push(e Event) {
	i := sort.Search(len(*l), func(i int) bool {
		return e.Timestamp.Before((*l)[i].Timestamp)
	})
	*l = append(*l, Event{})
	copy((*l)[i+1:], (*l)[i:])
	(*l)[i] = e
}

type ProjectProps struct {
	ID         uuid.UUID
	ActorProps []ProjectActorProps
}

type ProjectActorProps struct {
	PID       string
	NumEvents int32
	Alive     bool
}
