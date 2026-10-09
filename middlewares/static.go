package middlewares

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// StaticCacheMiddleware sets Cache-Control for static files.
// Versioned URLs (?v=<hash>) are cached forever, the rest must be revalidated on every request.
func StaticCacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.HasPrefix(c.Request().URL.Path, "/static/") {
			if c.QueryParam("v") != "" {
				c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000, immutable")
			} else {
				c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
			}
		}
		return next(c)
	}
}
