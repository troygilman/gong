package domain

import (
	"bytes"
	"context"
	"fmt"
	"gong/types"
	"io"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	ActorKindListener     = "listener"
	listenerTimerDuration = 50 * time.Millisecond
)

type listener struct {
	projectID uuid.UUID
	consumer  *actor.PID
	batch     map[string]types.EventList
	timerStop chan<- struct{}
}

func newListener(projectID uuid.UUID, consumer *actor.PID) actor.Producer {
	return func() actor.Receiver {
		return &listener{
			projectID: projectID,
			consumer:  consumer,
			batch:     make(map[string]types.EventList),
		}
	}
}

func (l *listener) Receive(ctx *actor.Context) {
	//actor_utils.LogActorMsg(ctx)
	if err := l.receive(ctx); err != nil {
		actor_utils.LogActorError(ctx, err)
		if ctx.Sender() != nil {
			ctx.Respond(err)
		}
	}
}

func (l *listener) receive(ctx *actor.Context) error {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		timerStop := make(chan struct{})
		l.timerStop = timerStop
		go func(engine *actor.Engine, pid *actor.PID) {
			timer := time.NewTimer(listenerTimerDuration)
			for {
				select {
				case <-timer.C:
					timer.Reset(listenerTimerDuration)
					engine.Send(pid, MsgListenerSendBatch{})
				case <-timerStop:
					return
				}
			}
		}(ctx.Engine(), ctx.PID())

	case actor.Stopped:
		close(l.timerStop)
		ctx.Send(l.consumer, MsgSse{})

	case MsgEvent:
		event := msg.Event
		events := l.batch[event.Event.PID]
		events.Push(event)
		l.batch[event.Event.PID] = events

	case MsgListenerSendBatch:
		msgs := make([]any, 0)
		for pid, events := range l.batch {
			var sseType string = types.SseTypeNewEvents
			var numEvents int32
			var actorAlive bool

			if len(events) > 0 {
				numEvents = events[len(events)-1].Index + 1
				actorAlive = isActorAlive(events[len(events)-1].MsgType)
				if events[0].Index == 0 {
					sseType = types.SseTypeNewActor
				}
			}

			switch sseType {
			case types.SseTypeNewActor:
				props := types.ProjectActorProps{
					PID:       pid,
					NumEvents: int32(len(events)),
					Alive:     actorAlive,
				}
				msgs = append(msgs, MsgSseNewActor{
					ActorProps: props,
				})
			case types.SseTypeNewEvents:
				msgs = append(msgs, MsgSseNewEvents{
					ActorPID:   pid,
					NumEvents:  numEvents,
					ActorAlive: actorAlive,
				})
			}
		}

		for pid := range l.batch {
			delete(l.batch, pid)
		}

		msgs = append(msgs, MsgSsePing{})

		ctx.Send(l.consumer, MsgSse{Msgs: msgs})
	}

	return nil
}

func writeServerSentEvent(ctx context.Context, w io.Writer, event string, data any) error {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("event: %s\n", event))

	var d string
	switch data := data.(type) {
	case templ.Component:
		buff := bytes.NewBuffer([]byte{})
		if err := data.Render(ctx, buff); err != nil {
			return err
		}
		d = strings.ReplaceAll(buff.String(), "\n", "")
	default:
		if data == nil {
			d = ""
		}
	}

	sb.WriteString(fmt.Sprintf("data: %s\n\n", d))
	if _, err := fmt.Fprint(w, sb.String()); err != nil {
		return err
	}

	return nil
}
