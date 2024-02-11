package middlewares

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/modules/constants"
)

func IsAuthenticated(c *echo.Context) (bool, int) {

	authSession, err := session.Get(constants.AUTH_SESSION_NAME, *c)

	if err == nil && authSession != nil {
		userID := authSession.Values[constants.AUTH_USER_KEY]
		if userID != nil {
			return true, userID.(int)
		}
	}

	return false, 0
}
