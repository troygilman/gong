package nested_components

import (
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/route"
)

func Route() gong.Route {
	childComponent := NewChildComponent()
	parentComponent := NewParentComponent(childComponent)
	return route.New("/", parentComponent)
}
