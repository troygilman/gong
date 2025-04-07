package main

import (
	"net/http"

	"github.com/troygilman/gong"
)

func main() {
	db := newUserDatabase()

	userComponent := gong.NewComponent(userView{
		db: db,
	})

	g := gong.New(http.NewServeMux()).Routes(
		gong.NewRoute("/", gong.NewComponent(homeView{})).WithRoutes(
			gong.NewRoute("users", gong.NewComponent(listView{
				db:            db,
				UserComponent: userComponent,
			})),
			gong.NewRoute("user/{name}", gong.NewComponent(testView{
				db:            db,
				UserComponent: userComponent,
			})),
		),
	)

	if err := http.ListenAndServe(":8080", g); err != nil {
		panic(err)
	}
}
