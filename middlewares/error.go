package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	httperrors "github.com/svachaj/sambar-wall/modules/http-errors"
	"github.com/svachaj/sambar-wall/utils"
)

// CustomHTTPErrorHandler handles HTTP errors and returns appropriate error pages.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	switch code {
	case http.StatusNotFound:
		utils.HTMLWithStatus(c, code, httperrors.ErrorPage(httperrors.NotFoundComponent()))

	default:
		utils.HTMLWithStatus(c, code, httperrors.ErrorPage(httperrors.InternalServerErrorComponent()))

	}
}
