package main

import (
	"net/http"

	"github.com/troygilman/gong/example/nested_components"
	"github.com/troygilman/gong/mux"
)

func main() {
	g := mux.New().Routes(nested_components.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
