package gong

import (
	"context"
	"fmt"
	"net/http"

	"github.com/troygilman/gong/internal/bind"
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
	return getContext(ctx).request
}

// LoaderData retrieves data loaded by a route's loader function.
// The generic type parameter specifies the expected type of the loaded data.
// Returns the zero value of the specified type if no loader data is available.
func LoaderData[Data any](ctx context.Context) (data Data) {
	return getContext(ctx).component.Loader(ctx).(Data)
}

// Redirect sends a redirect response to the client with the specified path.
// Uses HTTP status code 303 (See Other) for the redirect.
// Returns an error if the redirect fails.
func Redirect(ctx context.Context, path string) error {
	gCtx := getContext(ctx)
	gCtx.writer.Reset()
	http.Redirect(gCtx.writer, gCtx.request, path, http.StatusSeeOther)
	return nil
}

// ChildRoute retrieves the child route from the current context.
// This is useful when working with nested routes and needing to access
// the currently active child route within a parent component.
func ChildRoute(ctx context.Context) Route {
	return getContext(ctx).childRoute
}

func OutletID(ctx context.Context) string {
	gCtx := getContext(ctx)
	return "gong_" + gCtx.route.ID() + "_outlet"
}

func TargetID(ctx context.Context, id string) string {
	gCtx := getContext(ctx)
	prefix := "gong_" + gCtx.route.ID()
	if gCtx.componentID != "" {
		prefix += "_" + gCtx.componentID
	}
	if id != "" {
		prefix += "_" + id
	}
	return prefix
}

func TriggerAfterSwap(id string) string {
	return fmt.Sprintf("htmx:afterSwap[detail.target.id === '%s'] from:body", id)
}

func TriggerAfterSwapOOB(id string) string {
	return fmt.Sprintf("htmx:oobAfterSwap[detail.target.id === '%s'] from:body", id)
}

func ActionHeaders(ctx context.Context, headers ...string) string {
	return buildHeaders(append(gongHeaders(ctx, GongRequestTypeAction), headers...))
}
