package domain

import (
	"bytes"
	"context"
	"fmt"
	"gong/types"
	"gong/view"
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
	sse       chan<- string
	batch     map[string]types.EventList
	timerStop chan<- struct{}
}

func newListener(projectID uuid.UUID, sse chan<- string) actor.Producer {
	return func() actor.Receiver {
		return &listener{
			projectID: projectID,
			sse:       sse,
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

	case MsgEvent:
		event := msg.Event
		events := l.batch[event.Event.PID]
		events.Push(event)
		l.batch[event.Event.PID] = events

	case MsgListenerSendBatch:
		builder := &strings.Builder{}
		for pid, events := range l.batch {
			err := func() error {
				var sseType string = types.SseTypeNewEvents
				var numEvents int32
				var actorAlive bool

				if len(events) > 0 {
					numEvents = events[len(events)-1].Index + 1
					actorAlive = events[len(events)-1].MsgType != ActorStopped
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
					if err := writeServerSentEvent(ctx.Context(), builder, types.SseTypeNewActor, view.ProjectActorListItem(props, l.projectID)); err != nil {
						return err
					}
				case types.SseTypeNewEvents:
					if err := writeServerSentEvent(ctx.Context(), builder, types.SseTypeNewEvents+pid, view.ProjectActorEvents(events)); err != nil {
						return err
					}
					if err := writeServerSentEvent(ctx.Context(), builder, types.SseTypeUpdateActorNumEvents+pid, view.ProjectActorBadge(pid, numEvents, actorAlive)); err != nil {
						return err
					}
				case types.SseTypeDeleteActor:
					if err := writeServerSentEvent(ctx.Context(), builder, types.SseTypeDeleteActor+pid, nil); err != nil {
						return err
					}
				}

				return nil
			}()

			if err != nil {
				return fmt.Errorf("error while writing SSE: %s", err.Error())
			}

			delete(l.batch, pid)
		}

		if err := writeServerSentEvent(ctx.Context(), builder, "ping", nil); err != nil {
			return err
		}

		l.sse <- builder.String()
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
