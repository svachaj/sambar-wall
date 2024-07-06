package courses

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
)

func MapCoursesRoutes(e *echo.Echo, h ICoursesHandler) {

	e.GET("/", h.GetCoursesList)

	e.GET("/kurzy", h.GetCoursesList)

	e.GET("/prihlaska/:id", h.ApplicationFormPage, middlewares.AuthMiddleware)

}
