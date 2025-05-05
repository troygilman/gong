package nested_components

import (
	"github.com/troygilman/gong/route"
)

func Route() route.Builder {
	childComponent := NewChildComponent()
	parentComponent := NewParentComponent(childComponent)
	return route.New("/", parentComponent)
}
