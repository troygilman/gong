package main

import (
	"context"
	"log"

	"github.com/troygilman/gong/example/tabs"
	"github.com/troygilman/gong/hooks"
	"github.com/troygilman/gong/server"
)

func main() {
	svr := server.New(server.WithErrorHandler(func(ctx context.Context, err error) {
		log.Println(err)
		hooks.Header(ctx).Set("Hx-Reswap", "none")
	}))
	svr.Route(tabs.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
