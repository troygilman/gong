package main

import (
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/example/bulk_update"
)

func main() {
	g := gong.New(http.NewServeMux()).Routes(bulk_update.Route())
	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
