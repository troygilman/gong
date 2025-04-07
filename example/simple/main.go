package main

import (
	"net/http"

	"github.com/troygilman/gong"
)

func main() {
	simpleComponent := gong.NewComponent(SimpleComponent{})

	g := gong.New(http.NewServeMux()).Routes(
		gong.NewRoute("/", simpleComponent),
	)

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
