package processor

import (
	"context"
	"crypto/rand"
	gong_common "gong/common"
	"gong/core/common"
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
	common.CommonActorData
	actors   map[string]gong_common.Actor
	entropy  ulid.MonotonicReader
	updates  map[string]bool
	batch    []gong_common.Event
	repeater actor.SendRepeater
}

func NewProcessorProducerWithApp(app common.App) actor.Producer {
	return func() actor.Receiver {
		return &processor{
			CommonActorData: common.CommonActorData{
				App: app,
			},
		}
	}
}

func (p *processor) Receive(aCtx *actor.Context) {
	switch aCtx.Message().(type) {
	case common.MsgSdkEvents, common.MsgProcessorUpdateState:
	default:
		actor_utils.LogActorMsg(aCtx)
	}

	ctx, cancel := context.WithTimeout(aCtx.Context(), 1*time.Second)
	defer cancel()

	if err := p.receive(ctx, aCtx); err != nil {
		actor_utils.LogActorError(aCtx, err)
	}
}

func (p *processor) receive(ctx context.Context, aCtx *actor.Context) error {
	switch msg := aCtx.Message().(type) {
	case actor.Initialized:
		p.actors = make(map[string]gong_common.Actor)
		p.entropy = ulid.Monotonic(rand.Reader, 1)
		p.updates = make(map[string]bool)
		p.batch = make([]gong_common.Event, 0)
	case actor.Started:
		p.repeater = aCtx.SendRepeat(aCtx.PID(), common.MsgProcessorUpdateState{}, sendUpdatesInterval)
	case actor.Stopped:
		p.repeater.Stop()
		return p.processUpdates(ctx, aCtx)
	case common.MsgSdkEvents:
		slices.SortFunc(msg.Events, func(a, b sdk.Event) int {
			return a.Timestamp.Compare(b.Timestamp)
		})
		for _, event := range msg.Events {
			if err := p.handleEvent(event); err != nil {
				return err
			}
		}
	case common.MsgProcessorUpdateState:
		return p.processUpdates(ctx, aCtx)
	}
	return nil
}

func (p *processor) handleEvent(sdkEvent sdk.Event) error {
	eventID, err := ulid.New(ulid.Timestamp(sdkEvent.Timestamp), p.entropy)
	if err != nil {
		return err
	}

	event := gong_common.Event{
		ID:        eventID,
		ProjectID: p.Id,
		Event:     sdkEvent,
	}

	actorData, actorExists := p.actors[event.PID]
	if !actorExists {
		actorData = gong_common.Actor{
			PID:       event.PID,
			ProjectID: p.Id,
			NumEvents: 0,
			Alive:     false,
		}
	}

	event.Index = int32(actorData.NumEvents)

	actorData.NumEvents++
	if event.ID.Compare(actorData.LastEventID) > 0 {
		actorData.Alive = common.IsActorAlive(event.Type)
		actorData.LastEventID = event.ID
	}
	p.actors[event.PID] = actorData

	if _, ok := p.updates[event.PID]; !ok {
		p.updates[event.PID] = actorExists
	}

	p.batch = append(p.batch, event)

	return nil
}

func (p *processor) processUpdates(ctx context.Context, aCtx *actor.Context) error {
	defer func() {
		p.batch = make([]gong_common.Event, 0)
		p.updates = make(map[string]bool)
	}()

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

	if err := p.App.Database().CreateEvents(ctx, p.batch); err != nil {
		return err
	}

	projectPID := aCtx.Engine().Registry.GetPID(common.KindProject, p.Id.String())
	actors := make([]gong_common.Actor, 0)

	for pid := range p.updates {
		actors = append(actors, p.actors[pid])
	}

	if len(actors) > 0 {
		aCtx.Send(projectPID, common.MsgUpdateProject{
			Actors: actors,
			Events: p.batch,
		})
	}

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
