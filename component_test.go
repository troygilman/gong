package gong

import (
	"testing"

	"github.com/troygilman/gong/internal/assert"
)

func TestComponentRenderView(t *testing.T) {
	mock := testComponent{
		view: testTemplComponent{text: "view"},
	}

	component := NewComponent(mock)

	testRender(t, component, gongContext{}, "view")
}

func TestComponentRenderAction(t *testing.T) {
	mock := testComponent{
		action: testTemplComponent{text: "action"},
	}

	component := NewComponent(mock)

	testRender(t, component.Action(), gongContext{}, "action")
}

func TestComponentRenderAction_withLoader(t *testing.T) {
	mock := testComponent{
		action:     testLoaderTemplComponent{},
		loaderData: "action",
	}

	component := NewComponent(mock)

	testRender(t, component.Action(), gongContext{}, "action")
}

func TestComponentFind(t *testing.T) {
	mock := testComponent{}

	component := NewComponent(mock, ComponentWithID("mock"))

	foundComponent, ok := component.Find("mock")

	assert.Equals(t, true, ok)
	assert.Equals(t, component, foundComponent)
}

func TestComponentFind_withNestedComponent(t *testing.T) {
	child := NewComponent(testComponent{}, ComponentWithID("mock"))

	component := NewComponent(testParentComponent{Child: child}, ComponentWithID("parent"))

	foundComponent, ok := component.Find("parent_mock")

	assert.Equals(t, true, ok)
	assert.Equals(t, child, foundComponent)
}
