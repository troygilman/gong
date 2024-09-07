package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"stagebase/internal/common"
	"stagebase/internal/core"
	"stagebase/internal/transport"
	"stagebase/sdk"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestIntegration_Client(t *testing.T) {
	const (
		projectIdString    = "2db8b7bc-5237-4a01-bf7a-d0339da782d9"
		serverPort         = 8888
		numClients         = 2
		numEventsPerClient = 100000
	)

	// Mock
	controller := gomock.NewController(t)
	defer controller.Finish()

	db := common.NewMockDatabase(controller)

	db.EXPECT().QueryActors(
		gomock.Any(),
		uuid.MustParse(projectIdString),
	).Return(
		[]common.Actor{},
		nil,
	)

	countChan := make(chan int, 100)
	db.EXPECT().CreateEvents(
		gomock.Any(),
		gomock.AssignableToTypeOf([]common.Event{}),
	).DoAndReturn(func(_ context.Context, events []common.Event) error {
		countChan <- len(events)
		return nil
	}).AnyTimes()

	db.EXPECT().CreateActor(
		gomock.Any(),
		gomock.AssignableToTypeOf(common.Actor{}),
	).Return(nil).AnyTimes()

	db.EXPECT().UpdateActor(
		gomock.Any(),
		gomock.AssignableToTypeOf(common.Actor{}),
	).Return(nil).AnyTimes()

	// App
	os.Setenv(core.EnvAuthSecret, "test_secret")
	app, err := core.NewApp(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer app.Shutdown()

	// Server
	e, errChan := transport.RunServer(app, nil, serverPort)
	defer func() {
		fmt.Println("Shutting down server")
		e.Shutdown(context.Background())
		err := <-errChan
		if err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
		fmt.Println("Server shutdown successful")
	}()

	// Client Engine
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		t.Error(err)
		return
	}

	config := sdk.NewDefaultStagebaseConfig(projectIdString)
	config.Host = "localhost"
	config.Port = serverPort
	config.Scheme = "ws"
	config.QueueSize = 1000
	stagebasePID := engine.Spawn(sdk.NewStagebase(config), "stagebase", actor.WithID("primary"))

	start := time.Now()
	serverPID := engine.Spawn(newServer(numClients, numEventsPerClient), "server", actor.WithID("1"), actor.WithMiddleware(sdk.StagebaseMiddleware()))
	engine.Poison(serverPID).Wait()

	fmt.Println("Clients stopped")
	engine.Poison(stagebasePID).Wait()
	fmt.Printf("Finished in %fs\n", -time.Until(start).Seconds())

	totalCount := 0
	for count := range countChan {
		totalCount += count
		fmt.Printf("Count: %d\n", totalCount)
		if totalCount == numClients*(numEventsPerClient+3)+3 { // must include 3 actor framework msgs per actor
			break
		}
	}
}

type testMsg struct {
	id string
}

func (m testMsg) Log() (slog.Level, string, []any) {
	return slog.LevelInfo, "test", []any{"id", "jsdjsidjsjdi", "pid", actor.NewPID("hello", "123")}
}

type server struct {
	numClients         int
	numEventsPerClient int
	pids               []*actor.PID
}

func newServer(numClients, numEventsPerClient int) actor.Producer {
	return func() actor.Receiver {
		return &server{
			numClients:         numClients,
			numEventsPerClient: numEventsPerClient,
		}
	}
}

func (s *server) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Initialized:
		for i := range s.numClients {
			s.pids = append(s.pids, ctx.Engine().Spawn(
				newClient(),
				"client",
				actor.WithID(strconv.Itoa(i)),
				actor.WithMiddleware(sdk.StagebaseMiddleware()),
			))
		}
	case actor.Started:
		for range s.numEventsPerClient {
			for i := range s.numClients {
				ctx.Send(s.pids[i], testMsg{})
			}
		}
		fmt.Println("Finished sending msgs")
	case actor.Stopped:
		wg := &sync.WaitGroup{}
		for i := range s.numClients {
			ctx.Engine().Poison(s.pids[i], wg)
		}
		wg.Wait()
	}
}

type client struct {
	count int
}

func newClient() actor.Producer {
	return func() actor.Receiver {
		return &client{}
	}
}

func (c *client) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case testMsg:
		c.count++
	}
}
