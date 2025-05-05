package main

import (
	"github.com/troygilman/gong/component"
	"github.com/troygilman/gong/route"
	"github.com/troygilman/gong/server"
)

func main() {
	simpleComponent := component.New(SimpleComponent{})

	svr := server.New().Routes(
		route.New("/", simpleComponent),
	)

	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
