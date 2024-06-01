package domain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"gong/domain/persistance"
	"gong/sdk"
	"gong/types"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	actor_utils "github.com/troygilman0/utils"
)

const DefaultTimeout = 10 * time.Second

type App struct {
	secret        []byte
	engine        *actor.Engine
	pids          map[string]*actor.PID
	producerFuncs map[string]producerFunc
	db            persistance.Database
}

type producerFunc func(*App) actor.Producer

const (
	envAuthSecret = "AUTH_SECRET"
)

func NewApp() (*App, error) {
	app := new(App)

	app.db = persistance.NewPostgresDatabase()

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

	app.producerFuncs[ActorKindProject] = NewProjectProducerWithApp
	app.producerFuncs[ActorKindProcessor] = NewProcessorProducerWithApp

	return app, nil
}

var nilUUID = uuid.Nil.String()

func (app *App) pid(kind string, id string) (*actor.PID, error) {
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

func (app *App) PID(kind string, id string) (*actor.PID, error) {
	return app.pid(kind, id)
}

func (app *App) Engine() *actor.Engine {
	return app.engine
}

func (app *App) JWT(tokenString string) (jwt.RegisteredClaims, error) {
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

func (app *App) Actor(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
) (actor types.Actor, err error) {
	projectPID, err := app.pid(ActorKindProject, projectID.String())
	if err != nil {
		return actor, err
	}

	msg := MsgGetProjectActorProps{
		ActorPID: actorPID,
	}

	r := app.engine.Request(projectPID, msg, DefaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetProjectActor](r)
	if err != nil {
		return actor, err
	}

	return response.Actor, nil
}

func (app *App) ProjectProps(
	ctx context.Context,
	projectID uuid.UUID,
) (project types.ProjectProps, err error) {
	projectPID, err := app.pid(ActorKindProject, projectID.String())
	if err != nil {
		return project, err
	}

	r := app.engine.Request(projectPID, MsgGetProjectProps{}, DefaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetProjectProps](r)
	if err != nil {
		return project, err
	}

	return response.Props, nil
}

func (app *App) Events(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	eventID ulid.ULID,
	numEvents int,
	filter types.EventsFilter,
) (events types.EventList, err error) {
	projectPID, err := app.pid(ActorKindProject, projectID.String())
	if err != nil {
		return events, err
	}

	msg := MsgGetProjectEvents{
		ActorPID:  actorPID,
		EventID:   eventID,
		NumEvents: numEvents,
		Filter:    filter,
	}

	r := app.engine.Request(projectPID, msg, DefaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetProjectEvents](r)
	if err != nil {
		return events, err
	}

	return response.Events, nil
}

func (app *App) ProcessEvents(
	ctx context.Context,
	projectID uuid.UUID,
	events []sdk.Event,
) error {
	processorPID, err := app.pid(ActorKindProcessor, projectID.String())
	if err != nil {
		return err
	}

	app.engine.Send(processorPID, MsgSdkEvents{
		Events: events,
	})
	return nil
}

func (app *App) SourcesSinks(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	flowType string,
) (pids []string, err error) {
	projectPID, err := app.pid(ActorKindProject, projectID.String())
	if err != nil {
		return pids, err
	}

	msg := MsgGetActorSourcesSinks{
		ActorPID: actorPID,
		Flow:     flowType,
	}

	r := app.engine.Request(projectPID, msg, DefaultTimeout)
	response, err := actor_utils.HandleResponse[RspGetActorSourcesSinks](r)
	if err != nil {
		return pids, err
	}

	return response.SourcesSinks, nil
}
