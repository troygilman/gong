package main

import (
	"github.com/troygilman/gong/example/bulk_update"
	"github.com/troygilman/gong/server"
)

func main() {
	svr := server.New().Routes(bulk_update.Route())
	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
