package common

import (
	"io"
	"log/slog"
	"os"
)

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
