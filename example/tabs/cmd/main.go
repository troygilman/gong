package main

import (
	"context"
	"log"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/tabs"
)

func main() {
	svr := gong.NewServer(gong.ServerWithErrorHandler(func(ctx context.Context, err error) {
		log.Println(err)
		gong.Header(ctx).Set("Hx-Reswap", "none")
	}))
	svr.Route(tabs.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
