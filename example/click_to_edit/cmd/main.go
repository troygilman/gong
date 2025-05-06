package main

import (
	"github.com/troygilman/gong/example/click_to_edit"
	"github.com/troygilman/gong/server"
)

func main() {
	svr := server.New()
	svr.Route(click_to_edit.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
