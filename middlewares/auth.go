package middlewares

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/modules/constants"
)

func IsAuthenticated(c *echo.Context) (bool, string, int) {

	authSession, err := session.Get(constants.AUTH_SESSION_NAME, *c)

	if err == nil && authSession != nil {
		userName := authSession.Values[constants.AUTH_USER_USERNAME]
		userID := authSession.Values[constants.AUTH_USER_ID]
		if userID != nil && userName != nil {
			return true, userName.(string), userID.(int)
		}
	}

	return false, "", -1
}

// middleware to check if user is authenticated
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		if ok, _, _ := IsAuthenticated(&c); ok {
			return next(c)
		}

		returnUrl := c.Request().URL.Path
		if c.Request().URL.RawQuery != "" {
			returnUrl += "?" + c.Request().URL.RawQuery
		}

		// set return URL in session
		authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
		authSession.Values[constants.AUTH_RETURN_URL] = returnUrl
		authSession.Save(c.Request(), c.Response())

		c.Response().Header().Set("HX-Redirect", returnUrl)

		return c.Redirect(302, constants.ROUTE_LOGIN)
	}
}
