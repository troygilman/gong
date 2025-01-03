package main

import (
	"net/http"

	"github.com/troygilman0/gong"
)

func main() {
	db := newUserDatabase()

	g := gong.New(http.NewServeMux())

	g.Route("/", listView{
		db: db,
		ItemView: itemView{
			db: db,
		},
	}, func(r gong.Route) {})

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
