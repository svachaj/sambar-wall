package agreement

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/modules/constants"
)

func MapAgreementRoutes(e *echo.Echo, h IAgreementHandlers) {

	e.GET(constants.ROUTE_AGREEMENT_START_PAGE, h.AgreementStartPage)

	e.POST(constants.ROUTE_AGREEMENT_CHECK_EMAIL, h.CheckEmail)

}
