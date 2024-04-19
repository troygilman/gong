package handler

import (
	"gong/domain"
	"log"

	"github.com/labstack/echo/v4"
)

func MiddlewareLogRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			log.Printf("%s: %s", ctx.Request().Method, ctx.Request().RequestURI)
			return next(ctx)
		}
	}
}

func MiddlewareWithApp(app domain.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("app", app)
			return next(ctx)
		}
	}
}

func MiddlewareJWT(app domain.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tokenCookie, err := ctx.Cookie("auth-token")
			if err != nil {
				return err
			}
			claims, err := app.JWT(tokenCookie.Value)
			if err != nil {
				return err
			}
			ctx.Set("claims", claims)
			return next(ctx)
		}
	}
}
