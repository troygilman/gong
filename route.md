# Route Documentation

## Overview
The `Route` interface and its implementation `gongRoute` provide a routing system for the Gong framework. This system allows for hierarchical routing with components and supports both static and dynamic route handling.

## Interface Definition

### Route Interface
```go
type Route interface {
    Child(path string) Route
    Children() []Route
    Root() Route
    Path() string
    Component() Component
    Render(ctx context.Context, w io.Writer) error
}
```

## Methods

### Child(path string) Route
- **Purpose**: Retrieves a child route for a given path
- **Parameters**:
  - `path`: The path to find the child route for
- **Returns**: The child Route if found, or the default child if no specific match is found
- **Behavior**:
  - First checks for an exact match in the children map
  - If no match is found and a default child exists, returns the default child
  - Returns nil if no match is found and no default child exists

### Children() []Route
- **Purpose**: Returns all direct child routes
- **Returns**: A slice containing all child routes
- **Note**: This method returns only direct children, not the entire route tree

### Root() Route
- **Purpose**: Returns the root route of the routing tree
- **Returns**: The topmost Route in the hierarchy
- **Behavior**: Traverses up the parent chain until reaching the root

### Path() string
- **Purpose**: Returns the path segment for this route
- **Returns**: The path string associated with this route

### Component() Component
- **Purpose**: Returns the component associated with this route
- **Returns**: The Component that this route renders

### Render(ctx context.Context, w io.Writer) error
- **Purpose**: Renders the route's component
- **Parameters**:
  - `ctx`: The context containing rendering information
  - `w`: The writer to output the rendered content
- **Behavior**:
  - Sets the current route in the context
  - If an action is being performed, finds and renders the specific component
  - Otherwise, renders the route's main component
- **Returns**: An error if rendering fails

## Implementation Details

The `gongRoute` struct implements the Route interface and contains:
- `path`: The path segment for this route
- `component`: The component to render
- `children`: A map of child routes
- `defaultChild`: A fallback route when no specific match is found
- `parent`: Reference to the parent route

## Usage Example

```go
// Create a new route
route := &gongRoute{
    path: "/home",
    component: homeComponent,
}

// Add a child route
childRoute := &gongRoute{
    path: "dashboard",
    component: dashboardComponent,
    parent: route,
}
route.children["dashboard"] = childRoute

// Render the route
err := route.Render(ctx, w)
if err != nil {
    // Handle error
}
```

## Notes
- The routing system supports hierarchical paths
- Each route can have multiple children
- A default child route can be specified for fallback handling
- The system supports both static and dynamic component rendering 