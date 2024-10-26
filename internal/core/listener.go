package core

import (
	"log"
	"stagebase/internal/model"
	"stagebase/pkg/dolly"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
)

func NewListenerProducer(
	app App,
	projectID uuid.UUID,
	lType ListenerType,
	c chan<- []any,
	filter model.Filter,
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
	filter    model.Filter
}

func (l *listener) Receive(aCtx *actor.Context) {
	dolly.LogActorMsg(aCtx)
	if err := func() error {
		switch msg := aCtx.Message().(type) {
		case actor.Started:
			projectPID, err := l.app.PID(KindProject, l.projectID.String())
			if err != nil {
				return err
			}
			aCtx.Send(projectPID, listenStartMsg{lType: l.lType})

		case actor.Stopped:
			projectPID, err := l.app.PID(KindProject, l.projectID.String())
			if err != nil {
				return err
			}

			aCtx.Send(projectPID, listenStopMsg{lType: l.lType})
			close(l.c)

		case listenDataMsg:
			filteredData := make([]any, 0, len(msg.data))

			for _, data := range msg.data {
				switch l.lType {
				case ListenerTypeProject:
					filteredData = msg.data
				case ListenerTypeEvents:
					if event, ok := data.(model.Event); ok {
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
			log.Printf("%+v", filteredData)
		}
		return nil
	}(); err != nil {
		dolly.LogActorError(aCtx, err)
	}
}
