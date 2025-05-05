package main

import (
	"github.com/troygilman/gong/example/nested_components"
	"github.com/troygilman/gong/server"
)

func main() {
	svr := server.New().Routes(nested_components.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
