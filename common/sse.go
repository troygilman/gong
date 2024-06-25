package common

const (
	SseTypeNewActor               = "new-actor"
	SseTypeDeleteActor            = "delete-actor:"
	SseTypeUpdateActorNumEvents   = "update-actor-num-events:"
	SseTypeNewEvents              = "new-events"
	SseTypeUpdateMetricLatency    = "update-metric-latency"
	SseTypeUpdateMetricThroughput = "update-metric-throughput"
)

type SseNewActor struct {
	Actor Actor
}

type SseNewEvents struct {
	Actor Actor
}
