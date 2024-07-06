package courses

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/constants"
	"github.com/svachaj/sambar-wall/modules/courses/models"
	coursesTemplates "github.com/svachaj/sambar-wall/modules/courses/templates"
	"github.com/svachaj/sambar-wall/modules/toasts"
	"github.com/svachaj/sambar-wall/utils"
)

type ICoursesHandler interface {
	GetCoursesList(c echo.Context) error
	ApplicationFormPage(c echo.Context) error
	ProcessApplicationForm(c echo.Context) error
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

func (h *CoursesHandler) ProcessApplicationForm(c echo.Context) error {

	// validate form
	applicationForm := models.ApplicationFormModel()
	params, _ := c.FormParams()

	isValid := applicationForm.ValidateFields(params)

	if !isValid {
		applicationFormComponent := coursesTemplates.ApplicationForm(applicationForm, nil)
		return utils.HTML(c, applicationFormComponent)
	}

	c.Response().Header().Set("HX-Redirect", constants.ROUTE_HOME)

	applicationFormComponent := coursesTemplates.ApplicationForm(applicationForm, toasts.SuccessToast("Přihlášení proběhlo úspěšně."))
	return utils.HTML(c, applicationFormComponent)

}
