package core

import (
	"context"
	"errors"
	"fmt"
	"gong/internal/common"
	"gong/internal/persistance"
	"gong/sdk"
	"os"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	envAuthSecret  = "AUTH_SECRET"
	defaultTimeout = 10 * time.Second
)

type gongApp struct {
	secret        []byte
	engine        *actor.Engine
	pids          map[string]*actor.PID
	producerFuncs map[string]producerFunc
	db            persistance.Database
}

type producerFunc func(App) actor.Producer

func NewApp(db persistance.Database) (App, error) {
	app := new(gongApp)

	app.db = db

	secret, ok := os.LookupEnv(envAuthSecret)
	if !ok {
		return nil, fmt.Errorf("%s is not set", envAuthSecret)
	}
	app.secret = []byte(secret)

	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		return nil, err
	}
	app.engine = engine

	app.pids = make(map[string]*actor.PID)
	app.producerFuncs = make(map[string]producerFunc)

	app.producerFuncs[KindProject] = NewProjectProducerWithApp
	app.producerFuncs[KindProcessor] = NewProcessorProducerWithApp

	engine.SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case actor.Started:
			ctx.Engine().Subscribe(ctx.PID())
		case actor.EventLogger:
			//log.Println(msg.Log())
		default:
			_ = msg
		}
	}, "event-stream-listener")

	return app, nil
}

var nilUUID = uuid.Nil.String()

func (app *gongApp) PID(kind string, id string) (*actor.PID, error) {
	if id == nilUUID {
		return nil, errors.New("invalid actor id")
	}
	if pid, ok := app.pids[kind]; ok {
		return pid, nil
	}
	pid := app.engine.Registry.GetPID(kind, id)
	if pid == nil {
		uuid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		producer, ok := app.producerFuncs[kind]
		if !ok {
			return nil, errors.New("actor kind does not have a producer registered")
		}
		pid = app.engine.Spawn(producer(app), kind, actor.WithID(id), actor.WithMiddleware(ActorMiddlewareLoad(uuid)))
	}
	return pid, nil
}

func (app *gongApp) Database() persistance.Database {
	return app.db
}

func (app *gongApp) JWT(tokenString string) (jwt.RegisteredClaims, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return app.secret, nil
	})
	if err != nil {
		return claims, err
	}
	if !token.Valid {
		return claims, errors.New("invalid jwt")
	}
	return claims, nil
}

func (app *gongApp) Actor(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
) (actor common.Actor, err error) {
	projectPID, err := app.PID(KindProject, projectID.String())
	if err != nil {
		return actor, err
	}

	msg := MsgGetProjectActorProps{
		ActorPID: actorPID,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetProjectActor](r)
	if err != nil {
		return actor, err
	}

	return response.Actor, nil
}

func (app *gongApp) ProjectProps(
	ctx context.Context,
	projectID uuid.UUID,
) (project common.ProjectProps, err error) {
	projectPID, err := app.PID(KindProject, projectID.String())
	if err != nil {
		return project, err
	}

	r := app.engine.Request(projectPID, MsgGetProjectProps{}, defaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetProjectProps](r)
	if err != nil {
		return project, err
	}

	return response.Props, nil
}

func (app *gongApp) Events(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	timestamp time.Time,
	numEvents int,
	filter common.Filter,
) (events common.EventList, err error) {
	projectPID, err := app.PID(KindProject, projectID.String())
	if err != nil {
		return events, err
	}

	msg := MsgGetProjectEvents{
		ActorPID:  actorPID,
		Timestamp: timestamp,
		NumEvents: numEvents,
		Filter:    filter,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetProjectEvents](r)
	if err != nil {
		return events, err
	}

	return response.Events, nil
}

func (app *gongApp) ProcessEvents(
	ctx context.Context,
	projectID uuid.UUID,
	events []sdk.Event,
) error {
	processorPID, err := app.PID(KindProcessor, projectID.String())
	if err != nil {
		return err
	}

	app.engine.Send(processorPID, MsgSdkEvents{
		Events: events,
	})
	return nil
}

func (app *gongApp) SourcesSinks(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	sources bool,
	rng common.Range,
) (pids []string, err error) {
	projectPID, err := app.PID(KindProject, projectID.String())
	if err != nil {
		return pids, err
	}

	msg := MsgGetActorSourcesSinks{
		ActorPID: actorPID,
		Sources:  sources,
		Range:    rng,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetActorSourcesSinks](r)
	if err != nil {
		return pids, err
	}

	return response.SourcesSinks, nil
}

func (app *gongApp) SourcesSinksNodeInfo(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	nodePID string,
	source bool,
	rng common.Range,
) (info common.SourcesSinksNodeInfo, err error) {
	projectPID, err := app.PID(KindProject, projectID.String())
	if err != nil {
		return info, err
	}

	msg := MsgGetSourcesSinksNodeInfo{
		ActorPID: actorPID,
		NodePID:  nodePID,
		Source:   source,
		Range:    rng,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetSourcesSinksNodeInfo](r)
	if err != nil {
		return info, err
	}

	return response.Info, nil
}

func (app *gongApp) ListenStart(
	ctx context.Context,
	projectID uuid.UUID,
	id string,
	lType ListenerType,
	filter common.Filter,
) (<-chan []any, error) {
	c := make(chan []any, 100)
	app.engine.Spawn(NewListenerProducer(app, projectID, lType, c, filter), KindListener, actor.WithID(id))
	return c, nil
}

func (app *gongApp) ListenStop(
	ctx context.Context,
	id string,
) error {
	listenerPID, err := app.PID(KindListener, id)
	if err != nil {
		return err
	}
	app.engine.Poison(listenerPID)
	return nil
}
