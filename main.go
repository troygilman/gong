package main

import (
	"embed"
	"gong/api"
	"gong/core"
	"gong/example"
	"gong/persistance"
	"log"
	"os"
)

//go:embed static
var static embed.FS

//go:embed migration
var sqlMigrations embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-c" {
		example.RunExample(example.ExampleClient)
	} else {
		log.Fatal(run())
	}
}

func run() error {
	db, err := persistance.NewPostgresDatabase()
	if err != nil {
		return err
	}

	app, err := core.NewApp(db)
	if err != nil {
		return err
	}

	if err := api.RunServer(app, static); err != nil {
		return err
	}

	return nil
}
