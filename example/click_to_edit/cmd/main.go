package main

import (
	"net/http"

	"github.com/troygilman/gong/example/click_to_edit"
	"github.com/troygilman/gong/mux"
)

func main() {
	g := mux.New().Routes(click_to_edit.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
