package courses

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
	coursesTemplates "github.com/svachaj/sambar-wall/modules/courses/templates"
	"github.com/svachaj/sambar-wall/utils"
)

type ICoursesHandler interface {
	GetCoursesList(c echo.Context) error
	ApplicationFormPage(c echo.Context) error
}

type CoursesHandler struct {
	service ICoursesService
}

func NewCoursesHandler(svc ICoursesService) ICoursesHandler {
	return &CoursesHandler{service: svc}
}

func (h *CoursesHandler) GetCoursesList(c echo.Context) error {

	courses, err := h.service.GetCoursesList()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	isAuthenticated, _ := middlewares.IsAuthenticated(&c)

	coursesListComponent := coursesTemplates.CoursesList(courses, isAuthenticated)
	coursesPage := coursesTemplates.CoursesPage(coursesListComponent, isAuthenticated)

	return utils.HTML(c, coursesPage)
}

func (h *CoursesHandler) ApplicationFormPage(c echo.Context) error {

	id := c.Param("id")

	applicationForm := coursesTemplates.ApplicationFormPage(id)

	return utils.HTML(c, applicationForm)
}
