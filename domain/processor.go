package domain

import (
	"gong/sdk"
	"gong/types"
	"log"
	"math/rand"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	actor_utils "github.com/troygilman0/utils"
)

const (
	ActorKindProcessor  = "processor"
	sendUpdatesInterval = 100 * time.Millisecond
)

type processor struct {
	commonActorData
	actors   map[string]types.ProjectActorProps
	entropy  ulid.MonotonicReader
	updates  map[string]bool
	batch    []types.Event
	repeater actor.SendRepeater
}

func NewProcessorProducer() actor.Producer {
	return func() actor.Receiver {
		return &processor{}
	}
}

func NewProcessorProducerWithApp(app App) actor.Producer {
	return func() actor.Receiver {
		return &processor{
			commonActorData: commonActorData{
				app: app,
			},
		}
	}
}

func (p *processor) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case MsgSdkEvents, MsgSendUpdates:
	default:
		actor_utils.LogActorMsg(ctx)
	}

	if err := p.receive(ctx); err != nil {
		actor_utils.LogActorError(ctx, err)
	}
}

func (p *processor) receive(ctx *actor.Context) error {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		p.actors = make(map[string]types.ProjectActorProps)
		p.entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
		p.updates = make(map[string]bool)
		p.batch = make([]types.Event, 0)
	case actor.Started:
		p.repeater = ctx.SendRepeat(ctx.PID(), MsgSendUpdates{}, sendUpdatesInterval)
	case actor.Stopped:
		p.repeater.Stop()
		return p.processUpdates(ctx)
	case MsgSdkEvents:
		for _, event := range msg.Events {
			if err := p.handleEvent(event); err != nil {
				return err
			}
		}
	case MsgSendUpdates:
		return p.processUpdates(ctx)
	}
	return nil
}

func (p *processor) handleEvent(sdkEvent sdk.Event) error {
	eventID, err := ulid.New(ulid.Timestamp(sdkEvent.Timestamp), p.entropy)
	if err != nil {
		return err
	}

	event := types.Event{
		ID:        eventID,
		ProjectID: p.id,
		Event:     sdkEvent,
	}

	actorProps, actorExists := p.actors[event.PID]
	if !actorExists {
		actorProps = types.ProjectActorProps{
			PID:       event.PID,
			ProjectID: p.id,
			NumEvents: 0,
			Alive:     false,
		}
	}

	event.Index = int32(actorProps.NumEvents)

	actorProps.NumEvents++
	if event.ID.Compare(actorProps.LastEventID) > 0 {
		actorProps.Alive = isActorAlive(event.Type)
		actorProps.LastEventID = event.ID
	}
	p.actors[event.PID] = actorProps

	if _, ok := p.updates[event.PID]; !ok {
		p.updates[event.PID] = actorExists
	}

	p.batch = append(p.batch, event)

	return nil
}

func (p *processor) processUpdates(ctx *actor.Context) error {
	defer func() {
		p.batch = make([]types.Event, 0)
		p.updates = make(map[string]bool)
	}()

	projectPID := ctx.Engine().Registry.GetPID(ActorKindProject, p.id.String())
	props := make([]types.ProjectActorProps, 0)
	for pid := range p.updates {
		props = append(props, p.actors[pid])
	}
	if len(props) > 0 {
		ctx.Send(projectPID, MsgUpdateProjectActorProps{Props: props})
	}

	for pid, actorExists := range p.updates {
		var err error
		if actorExists {
			err = p.app.DB().UpdateActor(p.actors[pid])
		} else {
			err = p.app.DB().CreateActor(p.actors[pid])
		}
		if err != nil {
			return err
		}
	}

	start := time.Now()
	if err := p.app.DB().CreateEvents(p.batch); err != nil {
		return err
	}
	log.Printf("Create Events - %d events %d ms", len(p.batch), time.Since(start).Milliseconds())

	return nil
}

func (p *processor) load(id uuid.UUID) error {
	p.id = id

	actorProps, err := p.app.DB().QueryActors(id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *processor) create(id uuid.UUID) error {
	p.id = id
	return nil
}
