package domain

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const DefaultTimeout = time.Second

type App interface {
	Engine() *actor.Engine
	PID(kind string, id string) (*actor.PID, error)
	JWT(tokenString string) (jwt.RegisteredClaims, error)
	DB() Database
}

type defaultApp struct {
	secret        []byte
	engine        *actor.Engine
	pids          map[string]*actor.PID
	producerFuncs map[string]producerFunc
	db            Database
}

type producerFunc func(App) actor.Producer

const (
	envAuthSecret = "AUTH_SECRET"
)

func NewDefaultApp() (App, error) {
	app := new(defaultApp)

	app.db = NewPostgresDatabase()

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

func (app *defaultApp) PID(kind string, id string) (*actor.PID, error) {
	if id == nilUUID {
		return nil, errors.New("invalid actor id")
	}
	if pid, ok := app.pids[kind]; ok {
		return pid, nil
	}
	pid := app.Engine().Registry.GetPID(kind, id)
	if pid == nil {
		uuid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		producer, ok := app.producerFuncs[kind]
		if !ok {
			return nil, errors.New("actor kind does not have a producer registered")
		}
		pid = app.Engine().Spawn(producer(app), kind, actor.WithID(id), actor.WithMiddleware(ActorMiddlewareLoad(uuid)))
	}
	return pid, nil
}

func (app *defaultApp) Engine() *actor.Engine {
	return app.engine
}

func (app *defaultApp) JWT(tokenString string) (jwt.RegisteredClaims, error) {
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

func (app *defaultApp) DB() Database {
	return app.db
}
