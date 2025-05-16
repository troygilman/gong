package hook

import (
	"context"
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/internal/bind"
	"github.com/troygilman/gong/internal/gctx"
	"github.com/troygilman/gong/internal/util"
)

// Bind decodes form data from the current HTTP request into the provided destination.
// It processes both URL query parameters and form values submitted via POST or GET requests.
// The destination must be a pointer to a struct or map with appropriate "form" tags.
// Returns an error if the form parsing or binding fails.
func Bind(ctx context.Context, dest any) error {
	r := Request(ctx)
	if err := r.ParseForm(); err != nil {
		return err
	}
	return bind.Bind(r.Form, dest)
}

// FormValue retrieves the first value for the given form key from the current request.
// This is useful for accessing form data submitted via POST or GET requests.
func FormValue(ctx context.Context, key string) string {
	return Request(ctx).FormValue(key)
}

// PathParam retrieves the value of a path parameter from the current request.
// This is useful for accessing dynamic segments in the URL path.
func PathParam(ctx context.Context, key string) string {
	return Request(ctx).PathValue(key)
}

// QueryParam retrieves the value of a query parameter from the current request.
// This is useful for accessing data passed in the URL's query string.
func QueryParam(ctx context.Context, key string) string {
	return Request(ctx).URL.Query().Get(key)
}

// Request returns the current HTTP request object from the context.
// This provides access to all request properties and methods.
func Request(ctx context.Context) *http.Request {
	return gctx.GetContext(ctx).Request
}

// LoaderData retrieves data loaded by a route's loader function.
// The generic type parameter specifies the expected type of the loaded data.
// Returns the zero value of the specified type if no loader data is available.
func LoaderData[Data any](ctx context.Context) (data Data) {
	return gctx.GetContext(ctx).Component.Loader(ctx).(Data)
}

// Redirect sends a redirect response to the client with the specified path.
// Uses HTTP status code 303 (See Other) for the redirect.
// Returns an error if the redirect fails.
func Redirect(ctx context.Context, path string) error {
	gCtx := gctx.GetContext(ctx)
	gCtx.Writer.Reset()
	http.Redirect(gCtx.Writer, gCtx.Request, path, http.StatusSeeOther)
	return nil
}

// Header returns the HTTP header map that will be sent by the response.
// This is useful for adding or modifying response headers.
func Header(ctx context.Context) http.Header {
	return gctx.GetContext(ctx).Writer.Header()
}

// ChildRoute retrieves the child route from the current context.
// This is useful when working with nested routes and needing to access
// the currently active child route within a parent component.
func ChildRoute(ctx context.Context) gong.Route {
	gCtx := gctx.GetContext(ctx)
	return gCtx.Route.Child(gCtx.ChildRouteIndex)
}

// OutletID generates a unique ID for an outlet component based on the current route.
// This is used internally by the outlet component for proper targeting in HTMX.
func OutletID(ctx context.Context) string {
	gCtx := gctx.GetContext(ctx)
	return "gong_" + gCtx.CurrentRouteID + "_outlet"
}

// ComponentID generates a unique ID for a component based on the current route and component.
// This is used internally by components for proper targeting in HTMX.
func ComponentID(ctx context.Context) string {
	gCtx := gctx.GetContext(ctx)
	prefix := "gong_" + gCtx.CurrentRouteID
	if gCtx.ComponentID != "" {
		prefix += "_" + gCtx.ComponentID
	}
	return prefix
}

// ActionHeaders generates the HTMX header string for action requests.
// This includes the standard Gong action headers plus any additional headers provided.
func ActionHeaders(ctx context.Context, headers ...string) string {
	return util.BuildHeaders(append(util.GongHeaders(ctx, gong.GongRequestTypeAction), headers...))
}

// LinkHeaders generates the HTMX header string for link navigation requests.
// This includes the standard Gong link headers plus any additional headers provided.
func LinkHeaders(ctx context.Context, headers ...string) string {
	return util.BuildHeaders(append(util.GongHeaders(ctx, gong.GongRequestTypeLink), headers...))
}
