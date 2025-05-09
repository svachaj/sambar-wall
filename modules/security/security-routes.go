package security

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/constants"
)

// MapSecurityRoutes maps the security-related routes to their respective handlers.
// It sets up the following routes:
// - GET /login: handled by h.Login
// - POST /login/step1: handled by h.SignInStep1
// - POST /login/step2: handled by h.SignInStep2
// - GET /login/magic-link: handled by h.SignMeIn
// - GET /sign-out: handled by h.SignOut
// - GET /user/account: handled by h.UserAccountPage, protected by AuthMiddleware
//
// Parameters:
// - e: an instance of Echo framework
// - h: an implementation of ISecurityHandlers interface
func MapSecurityRoutes(e *echo.Echo, h ISecurityHandlers) {

	e.GET(constants.ROUTE_LOGIN, h.Login)

	e.POST(constants.ROUTE_LOGIN_STEP1, h.SignInStep1)

	e.POST(constants.ROUTE_LOGIN_STEP2, h.SignInStep2)

	e.GET(constants.ROUTE_LOGIN_MAGIC_LINK, h.SignMeIn)

	e.GET(constants.ROUTE_SIGN_OUT, h.SignOut)

	e.GET(constants.ROUTE_USER_ACCOUNT, h.UserAccountPage, middlewares.AuthMiddleware)

}
