package component

import (
	"testing"

	"github.com/troygilman/gong/internal/assert"
	"github.com/troygilman/gong/internal/gctx"
	"github.com/troygilman/gong/internal/test_util"
)

func TestComponentRenderView(t *testing.T) {
	mock := test_util.MockComponent{
		MockView: test_util.TextTemplComponent{Text: "view"},
	}

	component := New(mock)

	test_util.TestRender(t, component, gctx.Context{}, "view")
}

func TestComponentRenderAction(t *testing.T) {
	mock := test_util.MockComponent{
		MockAction: test_util.TextTemplComponent{Text: "action"},
	}

	component := New(mock)

	test_util.TestRender(t, component.Action(), gctx.Context{}, "action")
}

func TestComponentRenderAction_withLoader(t *testing.T) {
	mock := test_util.MockComponent{
		MockAction:     test_util.LoaderTemplComponent{},
		MockLoaderData: "action",
	}

	component := New(mock)

	test_util.TestRender(t, component.Action(), gctx.Context{}, "action")
}

func TestComponentFind(t *testing.T) {
	mock := test_util.MockComponent{}

	component := New(mock, WithID("mock"))

	foundComponent, ok := component.Find("mock")

	assert.Equals(t, true, ok)
	assert.Equals(t, component, foundComponent)
}

func TestComponentFind_withNestedComponent(t *testing.T) {
	child := New(test_util.MockComponent{}, WithID("mock"))

	component := New(test_util.ParentComponent{Child: child}, WithID("parent"))

	foundComponent, ok := component.Find("parent_mock")

	assert.Equals(t, true, ok)
	assert.Equals(t, child, foundComponent)
}
