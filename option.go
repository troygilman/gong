package gong

import (
	"context"

	"github.com/a-h/templ"
)

type componentConfig struct {
	loader   Loader
	id       string
	method   string
	trigger  string
	swap     string
	cssClass templ.CSSClass
}

type ComponentOption func(c componentConfig) componentConfig

func WithLoader(loader Loader) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.loader = loader
		return c
	}
}

func WithLoaderData(data any) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.loader = LoaderFunc(func(ctx context.Context) any {
			return data
		})
		return c
	}
}

func WithTrigger(trigger string) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.trigger = trigger
		return c
	}
}

func WithSwap(swap string) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.swap = swap
		return c
	}
}

func WithID(id string) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.id = id
		return c
	}
}

func WithMethod(method string) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.method = method
		return c
	}
}

func WithCSSClass(cssClass templ.CSSClass) ComponentOption {
	return func(c componentConfig) componentConfig {
		c.cssClass = cssClass
		return c
	}
}
