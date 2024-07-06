package courses

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/constants"
)

func MapCoursesRoutes(e *echo.Echo, h ICoursesHandler) {

	e.GET(constants.ROUTE_HOME, h.GetCoursesList)

	e.GET(constants.ROUTE_COURSES, h.GetCoursesList)

	e.GET(constants.ROUTE_COURSES_APPLAICATION_FORM_PAGE, h.ApplicationFormPage, middlewares.AuthMiddleware)

	e.POST(constants.ROUTE_COURSES_APPLAICATION_FORM, h.ProcessApplicationForm, middlewares.AuthMiddleware)

}
