package main

import (
	"net/http"

	"github.com/troygilman0/gong"
)

func main() {
	g := gong.New(http.NewServeMux())

	g.Route("/", listHandler{
		db: newUserDatabase(),
	}, func(r gong.Route) {
		// r.Route(":id", itemHandler{}, func(r gong.Route) {})
	})

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
