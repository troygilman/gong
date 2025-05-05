package main

import (
	"net/http"

	"github.com/troygilman/gong/example/tabs"
	"github.com/troygilman/gong/mux"
)

func main() {
	g := mux.New().Routes(tabs.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
