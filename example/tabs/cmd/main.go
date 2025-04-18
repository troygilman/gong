package main

import (
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/tabs"
)

func main() {
	g := gong.New(http.NewServeMux()).Routes(tabs.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
