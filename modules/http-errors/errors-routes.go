package httperrors

import (
	"github.com/labstack/echo/v4"
)

func MapErrorsRoutes(e *echo.Echo, h IErrorsHandler) {

	e.GET("/404", h.NotFound)
	e.GET("/500", h.InternalServerError)

}
