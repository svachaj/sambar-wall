package utils

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func HTML(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func HTMLWithStatus(c echo.Context, statusCode int, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().Writer.WriteHeader(statusCode)
	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func Classes(classNames ...string) string {
	return strings.Join(classNames, " ")
}

func ClassIf(condition bool, className string, elseClassName string) string {
	if condition {
		return className
	}
	return elseClassName
}
