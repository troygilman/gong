package core

import (
	"context"
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
	"gong/common"
)

type ListenerType uint

const (
	ListenerTypeProject ListenerType = iota
)

func newListenerProducer(
	app App,
	projectID uuid.UUID,
	lType ListenerType,
	c chan<- []any,
	filter *common.Filter,
) actor.Producer {
	return func() actor.Receiver {
		return &listener{
			app:       app,
			projectID: projectID,
			lType:     lType,
			c:         c,
			filter:    filter,
		}
	}
}

type listener struct {
	app       App
	projectID uuid.UUID
	lType     ListenerType
	c         chan<- []any
	filter    *common.Filter
}

func (l *listener) Receive(aCtx *actor.Context) {
	ctx := aCtx.Context()

	if err := l.receive(ctx, aCtx); err != nil {
		actor_utils.LogActorError(aCtx, err)
	}
}

func (l *listener) receive(ctx context.Context, aCtx *actor.Context) error {
	switch msg := aCtx.Message().(type) {
	case actor.Started:
		projectPID, err := l.app.pid(kindProject, l.projectID.String())
		if err != nil {
			return err
		}

		aCtx.Send(projectPID, msgListenStart{lType: l.lType})

	case actor.Stopped:
		projectPID, err := l.app.pid(kindProject, l.projectID.String())
		if err != nil {
			return err
		}

		aCtx.Send(projectPID, msgListenStop{lType: l.lType})
		close(l.c)

	case msgListenData:
		switch l.lType {
		case ListenerTypeProject:
			l.c <- msg.data
		default:
			panic("invalid listener type")
		}
	}
	return nil
}
