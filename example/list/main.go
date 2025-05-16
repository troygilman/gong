package main

import (
	"github.com/troygilman/gong"
)

func main() {
	db := newUserDatabase()

	userComponent := gong.NewComponent(userView{
		db: db,
	})

	svr := gong.NewServer()
	svr.Route(gong.NewRoute("/", gong.NewComponent(homeView{}),
		gong.WithChildren(
			gong.NewRoute("users", gong.NewComponent(listView{
				db:            db,
				UserComponent: userComponent,
			})),
			gong.NewRoute("user/{name}/", gong.NewComponent(testView{
				db:            db,
				UserComponent: userComponent,
			})),
		),
	))

	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
