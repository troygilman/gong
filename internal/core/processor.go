package core

import (
	"context"
	"gong/internal/common"
	"gong/pkg/dolly"
	"gong/sdk"
	"slices"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
)

const (
	sendUpdatesInterval = 50 * time.Millisecond
)

type processor struct {
	commonActorData
	actors   map[string]common.Actor
	updates  map[string]bool
	events   []sdk.Event
	repeater actor.SendRepeater
}

func NewProcessorProducerWithApp(app common.App) actor.Producer {
	return func() actor.Receiver {
		return &processor{
			commonActorData: commonActorData{
				app: app,
			},
		}
	}
}

func (p *processor) Receive(aCtx *actor.Context) {
	switch aCtx.Message().(type) {
	case msgSdkEvents, msgProcessorUpdateState:
	default:
		dolly.LogActorMsg(aCtx)
	}

	ctx, cancel := context.WithTimeout(aCtx.Context(), 10*time.Second)
	defer cancel()

	if err := func() error {
		switch msg := aCtx.Message().(type) {
		case actor.Initialized:
			p.actors = make(map[string]common.Actor)
			p.updates = make(map[string]bool)
			p.events = make([]sdk.Event, 0)
		case actor.Started:
			p.repeater = aCtx.SendRepeat(aCtx.PID(), msgProcessorUpdateState{}, sendUpdatesInterval)
		case actor.Stopped:
			p.repeater.Stop()
			return p.processUpdates(ctx, aCtx)
		case msgSdkEvents:
			slices.SortFunc(msg.events, func(a, b sdk.Event) int {
				return a.Timestamp.Compare(b.Timestamp)
			})
			p.events = append(p.events, msg.events...)
		case msgProcessorUpdateState:
			return p.processUpdates(ctx, aCtx)
		}
		return nil
	}(); err != nil {
		dolly.LogActorError(aCtx, err)
	}
}

func (p *processor) handleEvent(sdkEvent sdk.Event) (event common.Event, err error) {
	event = common.Event{
		ProjectID: p.id.String(),
		Event:     sdkEvent,
	}

	actorData, actorExists := p.actors[event.PID]
	if !actorExists {
		actorData = common.Actor{
			PID:       event.PID,
			ProjectID: p.id,
			NumEvents: 0,
			Alive:     false,
		}
	}

	actorData.NumEvents++

	if event.Timestamp.Compare(actorData.LastEventTime) > 0 {
		actorData.Alive = isActorAlive(event.Type)
		actorData.LastEventTime = event.Timestamp
	}

	p.actors[event.PID] = actorData

	if _, ok := p.updates[event.PID]; !ok {
		p.updates[event.PID] = actorExists
	}

	return event, nil
}

func (p *processor) processUpdates(ctx context.Context, aCtx *actor.Context) error {
	start := time.Now()
	batchSize := min(len(p.events), 10000)

	defer func() {
		p.events = p.events[batchSize:]
		p.updates = make(map[string]bool)
	}()

	sdkEventBatch := p.events[:batchSize]
	eventBatch := make([]common.Event, 0, len(sdkEventBatch))

	for _, sdkEvent := range sdkEventBatch {
		event, err := p.handleEvent(sdkEvent)
		if err != nil {
			continue
		}
		eventBatch = append(eventBatch, event)
	}

	if len(eventBatch) > 0 {
		if err := p.app.Database().CreateEvents(ctx, eventBatch); err != nil {
			return err
		}
	}

	for pid, actorExists := range p.updates {
		var err error
		if actorExists {
			err = p.app.Database().UpdateActor(ctx, p.actors[pid])
		} else {
			err = p.app.Database().CreateActor(ctx, p.actors[pid])
		}
		if err != nil {
			return err
		}
	}

	actors := make([]common.Actor, 0)
	for pid := range p.updates {
		actors = append(actors, p.actors[pid])
	}

	latency := time.Since(start)
	latencySec := float32(latency.Milliseconds()) / 1000

	var throughput int64
	if latencySec > 0 {
		throughput = int64(float32(batchSize) / latencySec)
	}

	metric := Metric{
		Timestamp:  time.Now(),
		Latency:    latency.Milliseconds(),
		Throughput: throughput,
	}

	projectPID := aCtx.Engine().Registry.GetPID(kindProject, p.id.String())
	aCtx.Send(projectPID, msgUpdateProject{
		actors: actors,
		events: eventBatch,
		metric: metric,
	})

	return nil
}

func (p *processor) Load(ctx context.Context, id uuid.UUID) error {
	p.id = id

	actorProps, err := p.app.Database().QueryActors(ctx, id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *processor) Create(ctx context.Context, id uuid.UUID) error {
	p.id = id
	return nil
}
