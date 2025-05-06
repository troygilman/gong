package main

import (
	"github.com/troygilman/gong/example/search"
	"github.com/troygilman/gong/server"
)

func main() {
	svr := server.New()
	svr.Route(search.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
