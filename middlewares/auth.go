package middlewares

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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

		returnUrl := c.Request().URL.Path
		if c.Request().URL.RawQuery != "" {
			returnUrl += "?" + c.Request().URL.RawQuery
		}

		log.Info().Msgf("User is not authenticated, redirecting to login page")
		log.Info().Msgf("Return URL: %s", returnUrl)

		// set return URL in session
		authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
		authSession.Values[constants.AUTH_RETURN_URL] = returnUrl
		authSession.Save(c.Request(), c.Response())

		return utils.HTMLWithStatus(c, 401, security.LoginPage())
	}
}
