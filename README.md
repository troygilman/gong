# Gong

A lightweight, component-based Go web framework with [templ](https://templ.guide) integration that makes building modern web applications simple and maintainable.

## Installation

```bash
go get github.com/troygilman/gong@latest
```

## Overview

Gong lets you build web applications in Go using components with encapsulated state and behavior:

- 🧩 Component-based architecture with HTMX integration
- 🔌 Automatic form binding and state management
- 🛣️ Comprehensive routing with path parameters
- 🧠 Context system for state propagation

## Example: Counter Component

```go
type CounterComponent struct {}

templ (c CounterComponent) View() {
	@gong.Target() {
		@counter(0)
	}
}

templ (c CounterComponent) Action() {
	{{
		count, err := strconv.Atoi(gong.FormValue(ctx, "count"))
		if err != nil {
			return err
		}
	}}
	@counter(count+1)
}

templ counter(count int) {
	<p>Count: { strconv.Itoa(count) }</p>
	@gong.Button() {
		Increment
		<input type="hidden" name="count" value={ strconv.Itoa(count) }/>
	}
}
```

## Documentation

For detailed documentation, examples, and guides, visit [https://gong-wiki.fly.dev](https://gong-wiki.fly.dev)

## Examples

Explore sample applications in the [example](example) directory.

## License

MIT License - see the LICENSE file for details.
