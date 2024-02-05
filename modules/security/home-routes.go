package security

import (
	"github.com/labstack/echo/v4"
)

func MapSecurityRoutes(e *echo.Echo, h ISecurityHandlers) {

	e.GET("/login/modal", h.LoginModal)

}
