package middlewares

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/config"
)

// session middleware with cookie store

func InitSessionMiddleware(settings *config.Config) echo.MiddlewareFunc {
	return session.Middleware(sessions.NewCookieStore([]byte(settings.AppSecret)))
}
