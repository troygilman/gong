package main

import (
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/click_to_edit"
)

func main() {
	g := gong.New(http.NewServeMux()).Routes(click_to_edit.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
