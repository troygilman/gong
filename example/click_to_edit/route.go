package click_to_edit

import "github.com/troygilman/gong"

func Route() gong.RouteBuilder {
	return gong.NewRoute("/", NewUserDetailComponent())
}
