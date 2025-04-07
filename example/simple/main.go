package main

import (
	"net/http"

	"github.com/troygilman/gong"
)

func main() {
	g := gong.New(http.NewServeMux()).Routes(
		gong.NewRoute("/", gong.NewComponent(SimpleComponent{})),
	)

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
