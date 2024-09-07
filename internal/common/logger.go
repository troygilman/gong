package common

import (
	"context"
	"io"
	"log/slog"
	"os"
)

func WithContextHandler(logger *slog.Logger, keys ...CtxKey) *slog.Logger {
	return slog.New(contextHandler{
		Handler: logger.Handler(),
		keys:    keys,
	})
}

type contextHandler struct {
	slog.Handler
	keys []CtxKey
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.observe(ctx)...)
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) observe(ctx context.Context) []slog.Attr {
	attrs := []slog.Attr{}
	for _, key := range h.keys {
		if val := ctx.Value(key); val != nil {
			attrs = append(attrs, slog.Any(string(key), val))
		}
	}
	return attrs
}

type loggerOption func(*loggerConfig)

func ConfigureDefaultLogger(options ...loggerOption) {
	config := newLoggerConfig()
	config.apply(options...)
	slog.SetDefault(slog.New(config.build()))
	slog.Default().Handler()
}

func newLoggerConfig() *loggerConfig {
	return &loggerConfig{
		json: false,
		w:    os.Stdout,
	}
}

type loggerConfig struct {
	json bool
	w    io.Writer
	opts *slog.HandlerOptions
}

func (config *loggerConfig) apply(options ...loggerOption) {
	for _, option := range options {
		option(config)
	}
}

func (config *loggerConfig) build() slog.Handler {
	if config.json {
		return slog.NewJSONHandler(config.w, config.opts)
	}
	return slog.NewTextHandler(config.w, config.opts)
}

func WithJSON() loggerOption {
	return func(config *loggerConfig) {
		config.json = true
	}
}

func WithOutput(w io.Writer) loggerOption {
	return func(config *loggerConfig) {
		config.w = w
	}
}

func WithHandlerOptions(opts *slog.HandlerOptions) loggerOption {
	return func(config *loggerConfig) {
		config.opts = opts
	}
}
