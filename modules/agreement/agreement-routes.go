package agreement

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/constants"
)

func MapAgreementRoutes(e *echo.Echo, h IAgreementHandlers) {

	e.GET(constants.ROUTE_AGREEMENT_START_PAGE, h.AgreementStartPage)

	e.POST(constants.ROUTE_AGREEMENT_CHECK_EMAIL, h.CheckEmail)

	e.POST(constants.ROUTE_AGREEMENT_FINALIZE, h.Finalize)

	// export application forms
	e.GET(constants.ROUTE_AGREEMENT_EXPORT_EMAILS_INIT, h.ExportEmailsConfirmedForCommercialCommunicationInit, middlewares.AuthRoleMiddleware(constants.ROLE_SAMBAR_ADMIN))
	e.GET(constants.ROUTE_AGREEMENT_EXPORT_EMAILS, h.ExportEmailsConfirmedForCommercialCommunication, middlewares.AuthRoleMiddleware(constants.ROLE_SAMBAR_ADMIN))

}
