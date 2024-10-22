package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"stagebase/internal/common"
	"stagebase/internal/model"
	"stagebase/internal/persistance"
	"stagebase/pkg/dolly"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	EnvAuthSecret  = "AUTH_SECRET"
	defaultTimeout = 10 * time.Second
)

type stagebaseApp struct {
	secret        []byte
	engine        *actor.Engine
	controllerPID *actor.PID
	pids          map[string]*actor.PID
	producerFuncs map[string]producerFunc
	db            persistance.Database
}

type producerFunc func(App) actor.Producer

func NewApp(db persistance.Database) (App, error) {
	app := new(stagebaseApp)

	app.db = db

	secret, ok := os.LookupEnv(EnvAuthSecret)
	if !ok {
		return nil, fmt.Errorf("%s is not set", EnvAuthSecret)
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

	app.controllerPID = engine.Spawn(newController(), kindController)

	return app, nil
}

func (app *stagebaseApp) Shutdown() {
	app.engine.Poison(app.controllerPID).Wait()
}

var nilUUID = uuid.Nil.String()

func (app *stagebaseApp) PID(kind string, id string) (*actor.PID, error) {
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
		pid = app.engine.Spawn(producer(app), kind, actor.WithID(id), actor.WithMiddleware(dolly.MiddlewareLoad(uuid)))
	}
	return pid, nil
}

func (app *stagebaseApp) Send(pid *actor.PID, msg any) {
	app.engine.Send(pid, msg)
}

func (app *stagebaseApp) Request(pid *actor.PID, msg any) (any, error) {
	response := app.engine.Request(pid, msg, 10*time.Second)
	result, err := response.Result()
	if err != nil {
		return nil, err
	}
	if err, ok := result.(error); ok {
		return nil, err
	}
	return result, nil
}

func (app *stagebaseApp) Database() persistance.Database {
	return app.db
}

func (app *stagebaseApp) JWT(tokenString string) (jwt.RegisteredClaims, error) {
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

func (app *stagebaseApp) SourcesSinks(
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

	msg := msgGetActorSourcesSinks{
		actorPID: actorPID,
		sources:  sources,
		rng:      rng,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := dolly.HandleResponse[rspGetActorSourcesSinks](r)
	if err != nil {
		return pids, err
	}

	return response.sourcesSinks, nil
}

func (app *stagebaseApp) SourcesSinksNodeInfo(
	ctx context.Context,
	projectID uuid.UUID,
	actorPID string,
	nodePID string,
	source bool,
	rng common.Range,
) (info model.SourcesSinksNodeInfo, err error) {
	projectPID, err := app.PID(KindProject, projectID.String())
	if err != nil {
		return info, err
	}

	msg := msgGetSourcesSinksNodeInfo{
		actorPID: actorPID,
		nodePID:  nodePID,
		source:   source,
		rng:      rng,
	}

	r := app.engine.Request(projectPID, msg, defaultTimeout)
	response, err := dolly.HandleResponse[rspGetSourcesSinksNodeInfo](r)
	if err != nil {
		return info, err
	}

	return response.info, nil
}

func (app *stagebaseApp) ListenStart(
	ctx context.Context,
	projectID uuid.UUID,
	id string,
	lType ListenerType,
	filter common.Filter,
) (<-chan []any, error) {
	c := make(chan []any, 100)
	app.engine.Spawn(NewListenerProducer(app, projectID, lType, c, filter), kindListener, actor.WithID(id))
	return c, nil
}

func (app *stagebaseApp) ListenStop(
	ctx context.Context,
	id string,
) error {
	listenerPID, err := app.PID(kindListener, id)
	if err != nil {
		return err
	}
	app.engine.Poison(listenerPID)
	return nil
}
