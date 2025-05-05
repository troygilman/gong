package main

import (
	"net/http"

	"github.com/troygilman/gong/component"
	"github.com/troygilman/gong/mux"
	"github.com/troygilman/gong/route"
)

func main() {
	db := newUserDatabase()

	userComponent := component.New(userView{
		db: db,
	})

	g := mux.New().Routes(
		route.New("/", component.New(homeView{})).WithRoutes(
			route.New("users", component.New(listView{
				db:            db,
				UserComponent: userComponent,
			})),
			route.New("user/{name}", component.New(testView{
				db:            db,
				UserComponent: userComponent,
			})),
		),
	)

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
