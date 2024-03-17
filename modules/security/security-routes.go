package security

import (
	"github.com/labstack/echo/v4"
)

func MapSecurityRoutes(e *echo.Echo, h ISecurityHandlers) {

	e.GET("/login/modal", h.LoginModal)

	e.POST("/sign-in-step1", h.SignInStep1)
	e.POST("/sign-in-step2", h.SignInStep2)

	e.GET("/sign-out", h.SignOut)

}
