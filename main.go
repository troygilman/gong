package main

import (
	"embed"
	"gong/example"
	"gong/internal/core"
	"gong/internal/persistance"
	"gong/internal/transport"
	"log"
	"os"
)

//go:embed static
var static embed.FS

//go:embed migration
var migrations embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-c" {
		example.RunExample(example.ExampleClient)
	} else {
		log.Fatal(run())
	}
}

func run() error {
	db, err := persistance.NewQuestDatabase(migrations)
	if err != nil {
		return err
	}

	app, err := core.NewApp(db)
	if err != nil {
		return err
	}

	if err := transport.RunServer(app, static); err != nil {
		return err
	}

	return nil
}
