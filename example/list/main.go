package main

import (
	"net/http"

	"github.com/troygilman/gong"
)

func main() {
	db := newUserDatabase()

	g := gong.New(http.NewServeMux())

	userComponent := gong.NewComponent("user", userView{
		db: db,
	})

	g.Route("/", homeView{}, func(r gong.Route) {
		r.Route("users", listView{
			db:            db,
			UserComponent: userComponent,
		}, nil)
		r.Route("user/{name}", testView{
			db:            db,
			UserComponent: userComponent,
		}, nil)
	})

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
