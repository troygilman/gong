package main

import (
	"net/http"

	"github.com/troygilman/gong"
)

func main() {
	db := newUserDatabase()

	userComponent := gong.NewComponent("user", userView{
		db: db,
	})

	g := gong.New(http.NewServeMux()).Routes(
		gong.NewRoute("/", homeView{}).WithRoutes(
			gong.NewRoute("users", listView{
				db:            db,
				UserComponent: userComponent,
			}),
			gong.NewRoute("user/{name}", testView{
				db:            db,
				UserComponent: userComponent,
			}),
		),
	)

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
