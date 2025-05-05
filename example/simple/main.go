package main

import (
	"net/http"

	"github.com/troygilman/gong/component"
	"github.com/troygilman/gong/mux"
	"github.com/troygilman/gong/route"
)

func main() {
	simpleComponent := component.New(SimpleComponent{})

	g := mux.New().Routes(
		route.New("/", simpleComponent),
	)

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
