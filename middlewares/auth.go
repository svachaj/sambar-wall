package middlewares

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/modules/constants"
)

// IsAuthenticated returns true if the user is authenticated, along with the username, user ID, and roles.
func IsAuthenticated(c *echo.Context) (bool, string, int, []string) {

	authSession, err := session.Get(constants.AUTH_SESSION_NAME, *c)

	if err == nil && authSession != nil {
		userName := authSession.Values[constants.AUTH_USER_USERNAME]
		userID := authSession.Values[constants.AUTH_USER_ID]
		roles := authSession.Values[constants.AUTH_USER_ROLES]
		if userID != nil && userName != nil && roles != nil {
			return true, userName.(string), userID.(int), roles.([]string)
		}
	}

	return false, "", -1, nil
}

// AuthMiddleware is a middleware to check if the user is authenticated.
// If the user is authenticated, it proceeds to the next handler.
// Otherwise, it redirects the user to the login route.
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		if ok, _, _, _ := IsAuthenticated(&c); ok {
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

// AuthRoleMiddleware is a middleware to check if the user is authenticated and has a specific role.
// If the user is authenticated and has the required role, it proceeds to the next handler.
// Otherwise, it redirects the user to the login route.
func AuthRoleMiddleware(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if ok, _, _, roles := IsAuthenticated(&c); ok && roles != nil {
				for _, r := range roles {
					if r == role {
						return next(c)
					}
				}
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
}

// HasRole checks if the given role is present in the list of roles.
func HasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// AuthMultiRoleMiddleware is a middleware to check if the user is authenticated and has any of the specified roles.
// If the user is authenticated and has at least one of the required roles, it proceeds to the next handler.
// Otherwise, it redirects the user to the login route.
func AuthMultiRoleMiddleware(allowedRoles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if ok, _, _, roles := IsAuthenticated(&c); ok && roles != nil {
				for _, userRole := range roles {
					for _, allowedRole := range allowedRoles {
						if userRole == allowedRole {
							return next(c)
						}
					}
				}
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
}
