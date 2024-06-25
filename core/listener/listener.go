package listener

import (
	"context"
	gong_common "gong/common"
	"gong/core/common"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

func NewListenerProducer(
	app common.App,
	projectID uuid.UUID,
	lType common.ListenerType,
	c chan<- []any,
	filter gong_common.Filter,
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
	app       common.App
	projectID uuid.UUID
	lType     common.ListenerType
	c         chan<- []any
	filter    gong_common.Filter
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
		projectPID, err := l.app.PID(common.KindProject, l.projectID.String())
		if err != nil {
			return err
		}

		aCtx.Send(projectPID, common.MsgListenStart{LType: l.lType})

	case actor.Stopped:
		projectPID, err := l.app.PID(common.KindProject, l.projectID.String())
		if err != nil {
			return err
		}

		aCtx.Send(projectPID, common.MsgListenStop{LType: l.lType})
		close(l.c)

	case common.MsgListenData:
		filteredData := make([]any, 0, len(msg.Data))

		for _, data := range msg.Data {
			switch l.lType {
			case common.ListenerTypeProject:
				filteredData = msg.Data
			case common.ListenerTypeEvents:
				if event, ok := data.(gong_common.Event); ok {
					if l.filter.ValidEvent(event) {
						filteredData = append(filteredData, event)
					}
				}
			case common.ListenerTypeMetrics:
				if metricGroup, ok := data.(common.MetricGroup); ok {
					filteredData = append(filteredData, metricGroup)
				}
			default:
				panic("invalid listener type")
			}
		}

		l.c <- filteredData
	}
	return nil
}
