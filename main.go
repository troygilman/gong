package main

import (
	"embed"
	"gong/api"
	"gong/core"
	"gong/example"
	"log"
	"os"
)

//go:embed static
var static embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-c" {
		example.RunExample(example.ExampleClient)
	} else {
		log.Fatal(run())
	}
}

func run() error {
	app, err := core.NewApp()
	if err != nil {
		return err
	}

	if err := api.RunServer(app, static); err != nil {
		return err
	}

	return nil
}
