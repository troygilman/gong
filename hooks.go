package gong

import (
	"context"
	"encoding/json"
	"net/http"
)

func Bind(ctx context.Context, dest any) error {
	gCtx := getContext(ctx)
	if err := json.NewDecoder(gCtx.request.Body).Decode(dest); err != nil {
		return err
	}
	return nil
}

func FormValue(ctx context.Context, key string) string {
	return Request(ctx).FormValue(key)
}

func Request(ctx context.Context) *http.Request {
	return getContext(ctx).request
}

func LoaderData[Data any](ctx context.Context) (data Data) {
	gCtx := getContext(ctx)
	if gCtx.loader == nil {
		return data
	}
	return gCtx.loader.Loader(ctx).(Data)
}

func Redirect(ctx context.Context, path string) error {
	gCtx := getContext(ctx)
	gCtx.writer.Reset()
	http.Redirect(gCtx.writer, gCtx.request, path, http.StatusSeeOther)
	return nil
}
