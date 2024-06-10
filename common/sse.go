package common

const (
	SseTypeNewActor             = "new-actor"
	SseTypeDeleteActor          = "delete-actor:"
	SseTypeUpdateActorNumEvents = "update-actor-num-events:"
	SseTypeNewEvents            = "new-events"
)

type SseNewActor struct {
	Actor Actor
}

type SseNewEvents struct {
	Actor Actor
}
