package gong

import "github.com/a-h/templ"

type elementConfig struct {
	id      string
	method  string
	swap    string
	target  string
	headers []string
	trigger string
	oob     bool
	attrs   templ.Attributes
	classes templ.CSSClasses
	node    *routeNode
}

type ElementOption func(elementConfig) elementConfig

func WithID(id string) ElementOption {
	return func(c elementConfig) elementConfig {
		c.id = id
		return c
	}
}

func WithMethod(method string) ElementOption {
	return func(c elementConfig) elementConfig {
		c.method = method
		return c
	}
}

func WithHeaders(headers ...string) ElementOption {
	return func(c elementConfig) elementConfig {
		c.headers = headers
		return c
	}
}

func WithAttrs(attrs templ.Attributes) ElementOption {
	return func(c elementConfig) elementConfig {
		c.attrs = attrs
		return c
	}
}

func WithSwap(swap string) ElementOption {
	return func(c elementConfig) elementConfig {
		c.swap = swap
		return c
	}
}

func WithClasses(classes ...any) ElementOption {
	return func(c elementConfig) elementConfig {
		c.classes = templ.Classes(classes...)
		return c
	}
}

func WithTarget(target string) ElementOption {
	return func(c elementConfig) elementConfig {
		c.target = target
		return c
	}
}

func WithTrigger(trigger string) ElementOption {
	return func(c elementConfig) elementConfig {
		c.trigger = trigger
		return c
	}
}

func withOOB(oob bool) ElementOption {
	return func(c elementConfig) elementConfig {
		c.oob = oob
		return c
	}
}

func withNode(node *routeNode) ElementOption {
	return func(c elementConfig) elementConfig {
		c.node = node
		return c
	}
}
