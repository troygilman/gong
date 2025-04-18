package main

import (
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/nested_components"
)

func main() {
	g := gong.New(http.NewServeMux()).Routes(nested_components.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
