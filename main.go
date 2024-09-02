package main

import (
	"embed"
	"log"
	"os"
	"stagebase/example"
	"stagebase/internal/common"
	"stagebase/internal/core"
	"stagebase/internal/persistance"
	"stagebase/internal/transport"
)

//go:embed static
var static embed.FS

//go:embed migrations
var migrations embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-c" {
		example.RunExample(example.ExampleClient)
	} else {
		log.Fatal(run())
	}
}

func run() error {
	common.ConfigureLogger()

	db, err := persistance.NewQuestDatabase(migrations)
	if err != nil {
		return err
	}

	app, err := core.NewApp(db)
	if err != nil {
		return err
	}

	_, errChan := transport.RunServer(app, static, 8080)
	err = <-errChan
	if err != nil {
		app.Shutdown()
		return err
	}

	return nil
}
