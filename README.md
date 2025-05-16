# Gong

Gong is a lightweight, component-based web framework for Go that integrates seamlessly with the [templ](https://templ.guide) templating system. It provides a simple and intuitive way to build modern web applications with Go, with a focus on developer experience and maintainable code.

## Features

- 🧩 **Component-based architecture** - Build reusable UI components with encapsulated state and behavior
- ⚡ **HTMX integration** - Built-in support for dynamic updates without writing JavaScript
- 🛣️ **Nested routing with path parameters** - Create complex navigation structures with ease
- 🔄 **Built-in action handling** - Handle user interactions and form submissions without boilerplate
- 📦 **Complete solution** - Everything you need to build modern web applications in one package
- 🔍 **Type safety** - Leverage Go's type system for safer, more maintainable code
- 🔌 **Form binding** - Automatically bind form data to Go structs with proper type conversion
- 🧠 **Context management** - Comprehensive context system for state propagation and request handling
- 📋 **Recursive form parsing** - Support for complex nested form structures with intuitive binding

## Installation

```bash
go get github.com/troygilman/gong@latest
```

## Quick Start

Here's a simple example of creating a counter component with Gong:

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
    	// Update state when the button is clicked
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

## Key Concepts

### Components

Components are the building blocks of your application. They encapsulate a view, actions, and optional data loading:

```go
// Create a component from any struct that implements View()
myComponent := component.New(MyViewStruct{})

// Add data loading capabilities
myComponent.WithLoaderFunc(func(ctx context.Context) any {
    return fetchData()
})

// Or set static data
myComponent.WithLoaderData(someData)

// Customize component ID for specific targeting
myComponent := component.New(MyViewStruct{}, component.WithID("my-custom-id"))
```

### Built-in UI Components

```go
// HTMX-powered buttons
button.New(
    button.WithMethod("POST"),
    button.WithSwap(gong.SwapOuterHTML),
    button.WithClass("btn btn-primary"),
    button.WithHeaders("X-Custom-Header", "value")
)

// Auto-updating targets
target.New(
    target.WithTrigger("load"),
    target.WithSwap(gong.SwapInnerHTML),
    target.WithAttrs(templ.Attributes{"aria-live": "polite"})
)

// Forms with automatic HTMX integration
form.New(
    form.WithMethod(http.MethodPost),
    form.WithSwap(gong.SwapInnerHTML),
    form.WithTarget("#result"),
    form.WithClass("form-container"),
    form.WithAttrs(templ.Attributes{"data-form-id": "user-form"})
)

// Client-side navigation links
link.New("/users",
    link.WithTrigger(gong.TriggerClick),
    link.WithID("user-link"),
    link.WithAttrs(templ.Attributes{"data-custom": "value"}),
    link.WithHeaders("X-Custom-Header", "value")
)
```

### Actions & State Management

Actions handle user interactions and update component state:

```go
templ (c MyComponent) Action() {
    // Access form values
    name := hooks.FormValue(ctx, "name")

    // Or bind entire form to a struct
    var user UserForm
    if err := hooks.Bind(ctx, &user); err == nil {
        // Form data is now available in user struct

        // Handles complex nested structures:
        // input name="user[name][first]" -> user.Name.First
        // input name="user[addresses][0][street]" -> user.Addresses[0].Street
        // input name="settings[theme]" -> settings["theme"]
    }

    // Update internal state
    c.someValue = name

    // Return updated UI
    <div>Hello, {name}!</div>
}
```

## Examples

### Simple Component

Check out the [simple example](example/simple) for a basic implementation.

### List Component

See the [list example](example/list) for a more complex implementation with nested routes and components.

### More Examples

- [Click to Edit](example/click_to_edit) - Implement inline editing with HTMX
- [Bulk Update](example/bulk_update) - Handle complex form submissions

## Context & Hooks

Gong provides a rich set of hooks to access context information and perform common operations:

```go
// Access HTTP request
req := hooks.Request(ctx)

// Get form values and query parameters
name := hooks.FormValue(ctx, "name")
id := hooks.QueryParam(ctx, "id")
param := hooks.PathParam(ctx, "user")

// Bind form data to a struct
var form UserForm
err := hooks.Bind(ctx, &form)

// Get data loaded by a component
user := hooks.LoaderData[User](ctx)

// Redirect to another page
hooks.Redirect(ctx, "/success")

// Access response headers
hooks.Header(ctx).Set("Custom-Header", "value")

// Work with routes
childRoute := hooks.ChildRoute(ctx)

// Get component IDs for client-side targeting
componentID := hooks.ComponentID(ctx)
outletID := hooks.OutletID(ctx)

// Generate HTMX headers for custom components
headers := hooks.ActionHeaders(ctx, "X-Custom", "value")
linkHeaders := hooks.LinkHeaders(ctx, "X-Navigation", "true")
```

The context system in Gong maintains all the necessary state during the request lifecycle, including request/response objects, component hierarchy, routing information, and action status.

## Documentation

Comprehensive documentation is available at [https://gong-wiki.fly.dev](https://gong-wiki.fly.dev).

## Component Architecture

Gong components are designed with a clean interface hierarchy:

```
Component
├── View - Renders the component's UI
├── Action - Handles user interactions
├── Loader - Fetches data for the component
└── Head - Provides head elements for the page
```

Each component can implement these interfaces to provide the necessary functionality. The component system uses Go's interface capabilities for extensibility while maintaining type safety.

Components maintain their own state and can find child components by ID, allowing for complex component trees with well-defined relationships:

```go
// Find a child component by ID
childComponent, found := parentComponent.Find("child-id")

// Composition allows components to be nested
type ParentComponent struct {
    ChildComponent gong.Component
}
```

Components can automatically discover their child components through reflection, creating a self-organizing component tree that mirrors your UI structure. This facilitates proper event bubbling and state propagation.

## Form Binding

Gong features a powerful form binding system that can handle complex nested structures:

```go
// Define a struct with form tags
type UserForm struct {
    Name struct {
        First string `form:"first"`
        Last  string `form:"last"`
    } `form:"name"`
    Addresses []struct {
        Street  string `form:"street"`
        City    string `form:"city"`
        Country string `form:"country"`
    } `form:"addresses"`
    Settings map[string]string `form:"settings"`
}

// HTML form with nested fields
// <input name="name[first]" value="John">
// <input name="name[last]" value="Doe">
// <input name="addresses[0][street]" value="123 Main St">
// <input name="settings[theme]" value="dark">

// Bind the form data
var form UserForm
err := hooks.Bind(ctx, &form)
```

The binding system supports various types including structs, maps, slices, primitive types, and types implementing `encoding.TextUnmarshaler`. It performs automatic type conversion and handles deeply nested structures through a specialized parser that efficiently manages memory allocations.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
