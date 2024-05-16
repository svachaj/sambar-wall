package middlewares

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/modules/constants"
	security "github.com/svachaj/sambar-wall/modules/security/templates"
	"github.com/svachaj/sambar-wall/utils"
)

func IsAuthenticated(c *echo.Context) (bool, string) {

	authSession, err := session.Get(constants.AUTH_SESSION_NAME, *c)

	if err == nil && authSession != nil {
		userID := authSession.Values[constants.AUTH_USER_KEY]
		if userID != nil {
			return true, userID.(string)
		}
	}

	return false, ""
}

// middleware to check if user is authenticated
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		if ok, _ := IsAuthenticated(&c); ok {
			return next(c)
		}

		return utils.HTMLWithStatus(c, 401, security.LoginPage())
	}
}
