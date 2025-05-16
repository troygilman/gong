package nested_components

import (
	"github.com/troygilman/gong"
)

func Route() gong.Route {
	childComponent := NewChildComponent()
	parentComponent := NewParentComponent(childComponent)
	return gong.NewRoute("/", parentComponent)
}
