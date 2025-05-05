package main

import (
	"github.com/troygilman/gong/component"
	"github.com/troygilman/gong/route"
	"github.com/troygilman/gong/server"
)

func main() {
	db := newUserDatabase()

	userComponent := component.New(userView{
		db: db,
	})

	svr := server.New().Routes(
		route.New("/", component.New(homeView{})).WithRoutes(
			route.New("users", component.New(listView{
				db:            db,
				UserComponent: userComponent,
			})),
			route.New("user/{name}/", component.New(testView{
				db:            db,
				UserComponent: userComponent,
			})),
		),
	)

	if err := svr.Run(":8080"); err != nil {
		panic(err)
	}
}
