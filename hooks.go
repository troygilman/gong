package gong

import (
	"context"
	"encoding/json"
	"net/http"
)

// Bind decodes the JSON body of the current HTTP request into the provided destination.
// The destination must be a pointer to a struct or map that matches the expected JSON structure.
// Returns an error if the JSON decoding fails.
func Bind(ctx context.Context, dest any) error {
	gCtx := getContext(ctx)
	if err := json.NewDecoder(gCtx.request.Body).Decode(dest); err != nil {
		return err
	}
	return nil
}

// FormValue retrieves the first value for the given form key from the current request.
// This is useful for accessing form data submitted via POST or GET requests.
func FormValue(ctx context.Context, key string) string {
	return Request(ctx).FormValue(key)
}

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
	gCtx := getContext(ctx)
	if gCtx.loader == nil {
		return data
	}
	return gCtx.loader.Loader(ctx).(Data)
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

func ChildRoute(ctx context.Context) Route {
	return getContext(ctx).childRoute
}
