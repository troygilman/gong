package domain

import (
	"container/heap"
	"gong/sdk"
	"log"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	ActorKindProject      = "Project"
	actorInitializedEvent = "actor.ActorInitializedEvent"
	actorStartedEvent     = "actor.ActorStartedEvent"
	actorStoppedEvent     = "actor.ActorStoppedEvent"
)

type project struct {
	commonActorData
	actors map[string]EventHeap
}

func NewProjectProducer() actor.Producer {
	return func() actor.Receiver {
		return &project{}
	}
}

func NewProjectProducerWithApp(app App) actor.Producer {
	return func() actor.Receiver {
		return &project{
			commonActorData: commonActorData{
				app: app,
			},
		}
	}
}

func (p *project) Receive(ctx *actor.Context) {
	actor_utils.LogActorMsg(ctx)
	if err := p.receive(ctx); err != nil {
		actor_utils.LogActorError(ctx, err)
		if ctx.Sender() != nil {
			ctx.Respond(err)
		}
	}
}

func (p *project) receive(ctx *actor.Context) error {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		p.actors = make(map[string]EventHeap)
	case MsgEvent:
		p.handleEvent(msg.Event)
	case MsgGetProjectProps:
		props := ProjectProps{
			Actors: make([]string, 0, len(p.actors)),
		}
		for actor := range p.actors {
			props.Actors = append(props.Actors, actor)
		}
		ctx.Respond(RspGetProjectProps{
			Props: props,
		})
	default:
		_ = msg
	}
	return nil
}

func (p *project) load(id uuid.UUID) error {
	p.id = id
	return nil
}

func (p *project) create(id uuid.UUID) error {
	p.id = id
	return nil
}

func (p *project) handleEvent(event sdk.Event) {
	actorPID := event.Receiver
	log.Printf("%+v", event)

	switch event.MsgType {
	case actorInitializedEvent, actorStartedEvent:
		pid, ok := parsePIDString(event.Attr["pid"])
		if !ok {
			break
		}
		actorPID = pid
	case actorStoppedEvent:
		pid, ok := parsePIDString(event.Attr["pid"])
		if !ok {
			break
		}
		delete(p.actors, pid)
		return
	}

	events, ok := p.actors[actorPID]
	if !ok {
		events = make(EventHeap, 0)
		heap.Init(&events)
	}

	heap.Push(&events, event)
	p.actors[actorPID] = events
	log.Println(events)
}

func parsePIDString(a any) (string, bool) {
	m, ok := a.(map[string]any)
	if !ok {
		return "", false
	}
	address, ok := m["address"].(string)
	if !ok {
		return "", false
	}
	id, ok := m["ID"].(string)
	if !ok {
		return "", false
	}
	return actor.NewPID(address, id).String(), true
}
