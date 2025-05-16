package main

import (
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/nested_components"
)

func main() {
	svr := gong.NewServer()
	svr.Route(nested_components.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
