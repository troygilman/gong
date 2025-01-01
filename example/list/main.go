package main

import (
	"net/http"

	"github.com/troygilman0/gong"
)

func main() {
	g := gong.New(http.NewServeMux())

	g.Route(gong.NewRoute("/", listHandler{
		ItemRoute: gong.NewRoute("item", itemHandler{}),
	}))

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
