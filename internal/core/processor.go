package core

import (
	"context"
	"crypto/rand"
	"gong/common"
	"gong/sdk"
	"slices"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	actor_utils "github.com/troygilman0/utils"
)

const (
	sendUpdatesInterval = 50 * time.Millisecond
)

type processor struct {
	CommonActorData
	actors   map[string]common.Actor
	entropy  ulid.MonotonicReader
	updates  map[string]bool
	events   []sdk.Event
	repeater actor.SendRepeater
}

func NewProcessorProducerWithApp(app App) actor.Producer {
	return func() actor.Receiver {
		return &processor{
			CommonActorData: CommonActorData{
				App: app,
			},
		}
	}
}

func (p *processor) Receive(aCtx *actor.Context) {
	switch aCtx.Message().(type) {
	case MsgSdkEvents, MsgProcessorUpdateState:
	default:
		actor_utils.LogActorMsg(aCtx)
	}

	ctx, cancel := context.WithTimeout(aCtx.Context(), 10*time.Second)
	defer cancel()

	if err := func() error {
		switch msg := aCtx.Message().(type) {
		case actor.Initialized:
			p.actors = make(map[string]common.Actor)
			p.entropy = ulid.Monotonic(rand.Reader, 1)
			p.updates = make(map[string]bool)
			p.events = make([]sdk.Event, 0)
		case actor.Started:
			p.repeater = aCtx.SendRepeat(aCtx.PID(), MsgProcessorUpdateState{}, sendUpdatesInterval)
		case actor.Stopped:
			p.repeater.Stop()
			return p.processUpdates(ctx, aCtx)
		case MsgSdkEvents:
			slices.SortFunc(msg.Events, func(a, b sdk.Event) int {
				return a.Timestamp.Compare(b.Timestamp)
			})
			p.events = append(p.events, msg.Events...)
		case MsgProcessorUpdateState:
			return p.processUpdates(ctx, aCtx)
		}
		return nil
	}(); err != nil {
		actor_utils.LogActorError(aCtx, err)
	}
}

func (p *processor) handleEvent(sdkEvent sdk.Event) (event common.Event, err error) {
	eventID, err := ulid.New(ulid.Timestamp(sdkEvent.Timestamp), p.entropy)
	if err != nil {
		return event, err
	}

	event = common.Event{
		ProjectID: p.Id.String(),
		Event:     sdkEvent,
	}

	actorData, actorExists := p.actors[event.PID]
	if !actorExists {
		actorData = common.Actor{
			PID:       event.PID,
			ProjectID: p.Id,
			NumEvents: 0,
			Alive:     false,
		}
	}

	actorData.NumEvents++

	lastEventID, err := ulid.Parse(actorData.LastEventID)
	if err != nil {
		lastEventID = ulid.ULID{}
	}

	if eventID.Compare(lastEventID) > 0 {
		actorData.Alive = IsActorAlive(event.Type)
		// actorData.LastEventID = event.ID
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

	if err := p.App.Database().CreateEvents(ctx, eventBatch); err != nil {
		return err
	}

	for pid, actorExists := range p.updates {
		var err error
		if actorExists {
			err = p.App.Database().UpdateActor(ctx, p.actors[pid])
		} else {
			err = p.App.Database().CreateActor(ctx, p.actors[pid])
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

	projectPID := aCtx.Engine().Registry.GetPID(KindProject, p.Id.String())
	aCtx.Send(projectPID, MsgUpdateProject{
		Actors: actors,
		Events: eventBatch,
		Metric: metric,
	})

	return nil
}

func (p *processor) Load(ctx context.Context, id uuid.UUID) error {
	p.Id = id

	actorProps, err := p.App.Database().QueryActors(ctx, id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *processor) Create(ctx context.Context, id uuid.UUID) error {
	p.Id = id
	return nil
}
