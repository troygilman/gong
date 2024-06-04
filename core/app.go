package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"gong/common"
	"gong/persistance"
	"gong/sdk"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	actor_utils "github.com/troygilman0/utils"
)

const (
	envAuthSecret  = "AUTH_SECRET"
	defaultTimeout = 10 * time.Second
)

type App interface {
	pid(kind string, id string) (*actor.PID, error)
	database() persistance.Database
	JWT(tokenString string) (jwt.RegisteredClaims, error)
	Actor(ctx context.Context, projectID uuid.UUID, actorPID string) (actor common.Actor, err error)
	ProjectProps(ctx context.Context, projectID uuid.UUID) (project common.ProjectProps, err error)
	Events(ctx context.Context, projectID uuid.UUID, actorPID string, eventID ulid.ULID, numEvents int, filter common.Filter) (events common.EventList, err error)
	ProcessEvents(ctx context.Context, projectID uuid.UUID, events []sdk.Event) error
	SourcesSinks(ctx context.Context, projectID uuid.UUID, actorPID string, flowType string) (pids []string, err error)
	ListenStart(ctx context.Context, projectID uuid.UUID, lType ListenerType, id string) (<-chan []any, error)
	ListenStop(ctx context.Context, id string) error
}

type gongApp struct {
	secret        []byte
	engine        *actor.Engine
	pids          map[string]*actor.PID
	producerFuncs map[string]producerFunc
	db            persistance.Database
}

type producerFunc func(App) actor.Producer

func NewApp() (App, error) {
	app := new(gongApp)

	db, err := persistance.NewPostgresDatabase()
	if err != nil {
		return nil, err
	}
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

	app.producerFuncs[kindProject] = newProjectProducerWithApp
	app.producerFuncs[kindProcessor] = newProcessorProducerWithApp

	return app, nil
}

var nilUUID = uuid.Nil.String()

func (app *gongApp) pid(kind string, id string) (*actor.PID, error) {
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
		pid = app.engine.Spawn(producer(app), kind, actor.WithID(id), actor.WithMiddleware(actorMiddlewareLoad(uuid)))
	}
	return pid, nil
}

func (app *gongApp) database() persistance.Database {
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
	projectPID, err := app.pid(kindProject, projectID.String())
	if err != nil {
		return actor, err
	}

	msg := msgGetProjectActorProps{
		ActorPID: actorPID,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := actor_utils.HandleResponse[rspGetProjectActor](r)
	if err != nil {
		return actor, err
	}

	return response.Actor, nil
}

func (app *gongApp) ProjectProps(
	ctx context.Context,
	projectID uuid.UUID,
) (project common.ProjectProps, err error) {
	projectPID, err := app.pid(kindProject, projectID.String())
	if err != nil {
		return project, err
	}

	r := app.engine.Request(projectPID, msgGetProjectProps{}, defaultTimeout)
	response, err := actor_utils.HandleResponse[rspGetProjectProps](r)
	if err != nil {
		return project, err
	}

	return response.Props, nil
}

func (app *gongApp) Events(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	eventID ulid.ULID,
	numEvents int,
	filter common.Filter,
) (events common.EventList, err error) {
	projectPID, err := app.pid(kindProject, projectID.String())
	if err != nil {
		return events, err
	}

	msg := msgGetProjectEvents{
		ActorPID:  actorPID,
		EventID:   eventID,
		NumEvents: numEvents,
		Filter:    filter,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := actor_utils.HandleResponse[rspGetProjectEvents](r)
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
	processorPID, err := app.pid(kindProcessor, projectID.String())
	if err != nil {
		return err
	}

	app.engine.Send(processorPID, msgSdkEvents{
		Events: events,
	})
	return nil
}

func (app *gongApp) SourcesSinks(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	flowType string,
) (pids []string, err error) {
	projectPID, err := app.pid(kindProject, projectID.String())
	if err != nil {
		return pids, err
	}

	msg := msgGetActorSourcesSinks{
		ActorPID: actorPID,
		Flow:     flowType,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := actor_utils.HandleResponse[rspGetActorSourcesSinks](r)
	if err != nil {
		return pids, err
	}

	return response.SourcesSinks, nil
}

func (app *gongApp) ListenStart(
	ctx context.Context,
	projectID uuid.UUID,
	lType ListenerType,
	id string,
) (<-chan []any, error) {
	c := make(chan []any, 100)
	app.engine.Spawn(newListenerProducer(app, projectID, lType, c, nil), kindListener, actor.WithID(id))
	return c, nil
}

func (app *gongApp) ListenStop(
	ctx context.Context,
	id string,
) error {
	listenerPID, err := app.pid(kindListener, id)
	if err != nil {
		return err
	}
	app.engine.Poison(listenerPID)
	return nil
}
