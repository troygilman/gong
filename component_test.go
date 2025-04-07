package gong

import (
	"testing"

	"github.com/troygilman/gong/assert"
)

func TestComponentRenderView(t *testing.T) {
	mock := MockComponent{
		view: textTemplComponent{"view"},
	}

	component := NewComponent("", mock)

	testRender(t, component, gongContext{}, "view")
}

func TestComponentRenderAction(t *testing.T) {
	mock := MockComponent{
		action: textTemplComponent{"action"},
	}

	component := NewComponent("", mock)

	testRender(t, component, gongContext{action: true}, "action")
}

func TestComponentRenderAction_withLoader(t *testing.T) {
	mock := MockComponent{
		action:     loaderTemplComponent{},
		loaderData: "action",
	}

	component := NewComponent("", mock)

	testRender(t, component, gongContext{action: true}, "action")
}

func TestComponentFind(t *testing.T) {
	mock := MockComponent{}

	component := NewComponent("", mock)

	foundComponent, ok := component.Find("")

	assert.Equals(t, true, ok)
	assert.Equals(t, component, foundComponent)
}

func TestComponentFind_withNestedComponent(t *testing.T) {
	child := NewComponent("child", MockComponent{})

	component := NewComponent("", ParentComponent{
		Child: child,
	})

	foundComponent, ok := component.Find("child")

	assert.Equals(t, true, ok)
	assert.Equals(t, child, foundComponent)
}
