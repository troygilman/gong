package main

import (
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/bulk_update"
)

func main() {
	svr := gong.NewServer()
	svr.Route(bulk_update.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
