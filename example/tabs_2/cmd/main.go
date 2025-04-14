package main

import (
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/tabs_2"
)

func main() {
	g := gong.New(http.NewServeMux()).Routes(tabs_2.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
