package main

import (
	"net/http"

	"github.com/troygilman/gong/example/bulk_update"
	"github.com/troygilman/gong/mux"
)

func main() {
	g := mux.New().Routes(bulk_update.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
