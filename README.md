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
package main

import (
    "github.com/troygilman/gong/button"
    "github.com/troygilman/gong/component"
    "github.com/troygilman/gong/route"
    "github.com/troygilman/gong/server"
)

// Define a counter component
type CounterComponent struct {}

// View method renders the component UI
templ (c CounterComponent) View() {
	@target.New() {
		<p>Count: 0</p>
		@button.New() {
			Increment
			<input type="hidden" name="count" value="0"/>
		}
	}
}

// Action method handles user interactions
templ (c CounterComponent) Action() {
    {{
     	count, err := strconv.Atoi(hooks.FormValue(ctx, "count"))
      	if err != nil {
       		return err
       	}
    }}
    <p>Count: {count}</p>
    @button.New() {
    	Increment
     	<input type="hidden" name="count" value={ strconv.Itoa(count+1) }/>
    }
}

func main() {
    // Create a Gong server
    svr := server.New()

    // Create a component instance
    counterComponent := component.New(CounterComponent{count: 0})

    // Set up routing
    svr.Route(route.New("/", counterComponent))

    // Start the server
    if err := svr.Run(":8080"); err != nil {
        panic(err)
    }
}
```

## Documentation

For detailed documentation, examples, and guides, visit [https://gong-wiki.fly.dev](https://gong-wiki.fly.dev)

## Examples

Explore sample applications in the [example](example) directory.

## License

MIT License - see the LICENSE file for details.