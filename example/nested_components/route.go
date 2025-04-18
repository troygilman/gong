package nested_components

import (
	"github.com/troygilman/gong"
)

func Route() gong.RouteBuilder {
	childComponent := NewChildComponent()
	parentComponent := NewParentComponent(childComponent)
	return gong.NewRoute("/", parentComponent)
}
