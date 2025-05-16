package main

import (
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/search"
)

func main() {
	svr := gong.NewServer()
	svr.Route(search.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
