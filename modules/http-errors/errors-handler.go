package httperrors

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/utils"
)

type IErrorsHandler interface {
	NotFound(c echo.Context) error
	InternalServerError(c echo.Context) error
}

type ErrorsHandler struct {
}

func NewErrorsHandler() IErrorsHandler {
	return &ErrorsHandler{}
}

func (h *ErrorsHandler) NotFound(c echo.Context) error {
	notFoundCmp := NotFoundComponent()
	pageNotFound := ErrorPage(notFoundCmp)

	return utils.HTMLWithStatus(c, 404, pageNotFound)
}

func (h *ErrorsHandler) InternalServerError(c echo.Context) error {
	internalServerCmp := InternalServerErrorComponent()
	pageInternalServerError := ErrorPage(internalServerCmp)

	return utils.HTMLWithStatus(c, 500, pageInternalServerError)
}
