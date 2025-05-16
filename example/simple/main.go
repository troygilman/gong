package main

import (
	"github.com/troygilman/gong"
)

func main() {
	simpleComponent := gong.NewComponent(SimpleComponent{})

	svr := gong.NewServer()
	svr.Route(gong.NewRoute("/", simpleComponent))

	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
