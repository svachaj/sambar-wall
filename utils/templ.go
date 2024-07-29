package utils

import (
	"strconv"
	"strings"
	"time"

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

func StringFromInt(i int) string {
	return strconv.Itoa(i)
}

func StringFromBool(b bool) string {
	return strconv.FormatBool(b)
}

func StringFromBoolForCheckbox(b bool) string {
	if b {
		return "checked"
	}
	return ""
}

func StringFromFloat64(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func StringFromFloat32(f float32) string {
	return strconv.FormatFloat(float64(f), 'f', -1, 32)
}

func StringFromInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func StringFromInt32(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

func StringFromInt16(i int16) string {
	return strconv.FormatInt(int64(i), 10)
}

func StringFromStringPointer(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StringFromBoolPointer(b *bool) string {
	if b == nil {
		return ""
	}
	return strconv.FormatBool(*b)
}

func StringFromIntPointer(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}

func StringFromDateTimePointer(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2.1.2006")
}

func StringFromDateTime(t time.Time) string {
	return t.Format("2.1.2006")
}

func StringifyBool(value bool) string {
	if value {
		return "true"
	} else {
		return "false"
	}
}
