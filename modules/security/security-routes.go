package security

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/constants"
)

func MapSecurityRoutes(e *echo.Echo, h ISecurityHandlers) {

	e.GET(constants.ROUTE_LOGIN, h.Login)

	e.POST(constants.ROUTE_LOGIN_STEP1, h.SignInStep1)

	e.POST(constants.ROUTE_LOGIN_STEP2, h.SignInStep2)

	e.GET(constants.ROUTE_LOGIN_MAGIC_LINK, h.SignMeIn)

	e.GET(constants.ROUTE_SIGN_OUT, h.SignOut)

	e.GET(constants.ROUTE_USER_ACCOUNT, h.UserAccountPage, middlewares.AuthMiddleware)

}
