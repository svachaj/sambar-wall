package home

import (
	"github.com/labstack/echo/v4"
)

func MapHomeRoutes(e *echo.Echo, h IHomeHandlers) {

	e.GET("/", h.Home)

}
