package main

import (
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/counter"
)

func main() {
	server := gong.NewServer()
	server.Route(gong.NewRoute("/", gong.NewComponent(counter.CounterComponent{})))
	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
