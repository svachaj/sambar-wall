package agreement

import (
	"github.com/labstack/echo/v4"
)

func MapAgreementRoutes(e *echo.Echo, h IAgreementHandlers) {

	e.GET("/souhlas-s-provoznim-radem", h.Step1)

}
