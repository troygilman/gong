package core

import (
	"gong/internal/common"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

func NewListenerProducer(
	app App,
	projectID uuid.UUID,
	lType ListenerType,
	c chan<- []any,
	filter common.Filter,
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
	filter    common.Filter
}

func (l *listener) Receive(aCtx *actor.Context) {
	if err := func() error {
		switch msg := aCtx.Message().(type) {
		case actor.Started:
			projectPID, err := l.app.PID(KindProject, l.projectID.String())
			if err != nil {
				return err
			}
			aCtx.Send(projectPID, MsgListenStart{LType: l.lType})

		case actor.Stopped:
			projectPID, err := l.app.PID(KindProject, l.projectID.String())
			if err != nil {
				return err
			}

			aCtx.Send(projectPID, MsgListenStop{LType: l.lType})
			close(l.c)

		case MsgListenData:
			filteredData := make([]any, 0, len(msg.Data))

			for _, data := range msg.Data {
				switch l.lType {
				case ListenerTypeProject:
					filteredData = msg.Data
				case ListenerTypeEvents:
					if event, ok := data.(common.Event); ok {
						if l.filter.ValidEvent(event) {
							filteredData = append(filteredData, event)
						}
					}
				case ListenerTypeMetrics:
					if metricGroup, ok := data.(MetricGroup); ok {
						filteredData = append(filteredData, metricGroup)
					}
				default:
					panic("invalid listener type")
				}
			}

			l.c <- filteredData
		}
		return nil
	}(); err != nil {
		actor_utils.LogActorError(aCtx, err)
	}
}
