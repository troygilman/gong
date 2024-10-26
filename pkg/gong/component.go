package gong

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type ComponentProducer func(ctx context.Context) (templ.Component, error)

func NewTemplHandler(producer ComponentProducer) echo.HandlerFunc {
	return func(e echo.Context) error {
		component, err := producer(e.Request().Context())
		if err != nil {
			return err
		}
		return component.Render(e.Request().Context(), e.Response())
	}
}

func NewComponent(producer ComponentProducer) templ.Component {
	return producerComponent{
		producer: producer,
	}
}

type producerComponent struct {
	producer ComponentProducer
}

func (c producerComponent) Render(ctx context.Context, w io.Writer) error {
	component, err := c.producer(ctx)
	if err != nil {
		return err
	}
	return component.Render(ctx, w)
}
