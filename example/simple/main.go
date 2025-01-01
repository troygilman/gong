package main

import (
	"net/http"

	"github.com/troygilman0/gong"
)

func main() {
	g := gong.New(http.NewServeMux())

	g.Handle("/home", homeHandler{})

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
