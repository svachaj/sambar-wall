package courses

import (
	"github.com/labstack/echo/v4"
)

func MapCoursesRoutes(e *echo.Echo, h ICoursesHandler) {

	e.GET("/", h.GetCoursesList)

	e.GET("/kurzy", h.GetCoursesList)

}
