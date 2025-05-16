package main

import (
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/click_to_edit"
)

func main() {
	svr := gong.NewServer()
	svr.Route(click_to_edit.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
