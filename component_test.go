package gong

import (
	"testing"

	"github.com/troygilman/gong/internal/assert"
)

func TestComponentRenderView(t *testing.T) {
	mock := MockComponent{
		MockView: TextTemplComponent{Text: "view"},
	}

	component := NewComponent(mock)

	TestRender(t, component, gongContext{}, "view")
}

func TestComponentRenderAction(t *testing.T) {
	mock := MockComponent{
		MockAction: TextTemplComponent{Text: "action"},
	}

	component := NewComponent(mock)

	TestRender(t, component.Action(), gongContext{}, "action")
}

func TestComponentRenderAction_withLoader(t *testing.T) {
	mock := MockComponent{
		MockAction:     LoaderTemplComponent{},
		MockLoaderData: "action",
	}

	component := NewComponent(mock)

	TestRender(t, component.Action(), gongContext{}, "action")
}

func TestComponentFind(t *testing.T) {
	mock := MockComponent{}

	component := NewComponent(mock, ComponentWithID("mock"))

	foundComponent, ok := component.Find("mock")

	assert.Equals(t, true, ok)
	assert.Equals(t, component, foundComponent)
}

func TestComponentFind_withNestedComponent(t *testing.T) {
	child := NewComponent(MockComponent{}, ComponentWithID("mock"))

	component := NewComponent(ParentComponent{Child: child}, ComponentWithID("parent"))

	foundComponent, ok := component.Find("parent_mock")

	assert.Equals(t, true, ok)
	assert.Equals(t, child, foundComponent)
}
