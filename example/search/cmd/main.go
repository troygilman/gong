package main

import (
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/search"
)

func main() {
	g := gong.New(http.NewServeMux()).Routes(search.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
