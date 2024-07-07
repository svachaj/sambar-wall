package courses

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/constants"
	"github.com/svachaj/sambar-wall/modules/courses/models"
	coursesTemplates "github.com/svachaj/sambar-wall/modules/courses/templates"
	httperrors "github.com/svachaj/sambar-wall/modules/http-errors"
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

	isAuthenticated, _, _ := middlewares.IsAuthenticated(&c)

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
	applicationForm := models.ApplicationFormModel(c.Param(models.APPLICATION_FORM_COURSE_ID))
	params, _ := c.FormParams()

	isValid := applicationForm.ValidateFields(params)

	if !isValid {
		applicationFormComponent := coursesTemplates.ApplicationForm(applicationForm, nil)
		return utils.HTML(c, applicationFormComponent)
	}

	// get values from the form
	firstName := applicationForm.FormFields[models.APPLICATION_FORM_FIRST_NAME].Value
	lastName := applicationForm.FormFields[models.APPLICATION_FORM_LAST_NAME].Value
	phone := applicationForm.FormFields[models.APPLICATION_FORM_PHONE].Value
	parentName := applicationForm.FormFields[models.APPLICATION_FORM_PARENT_NAME].Value

	courseId, err := strconv.Atoi(applicationForm.FormFields[models.APPLICATION_FORM_COURSE_ID].Value)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert courseId to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	personalIdString := applicationForm.FormFields[models.APPLICATION_FORM_PERSONAL_ID].Value
	personalId, err := strconv.Atoi(personalIdString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert personalId to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	// get username from the session
	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	userEmail := authSession.Values[constants.AUTH_USER_USERNAME].(string)
	userId := authSession.Values[constants.AUTH_USER_ID].(int)

	// check if the provided personalId and courseId isnt already in the database
	exists, err := h.service.CheckApplicationFormExists(courseId, personalId)
	if err != nil {
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}
	if exists {
		return utils.HTML(c, coursesTemplates.ApplicationFormErrorInfo("Tento účastník je již na tento kurz přihlášen."))
	}

	// create or use existing participant by first name, last name and birth year extracted from the personalId
	//get birth year from the personalIdString
	birthYear2 := personalIdString[0:2]
	birthYear2Int, err := strconv.Atoi(birthYear2)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert birthYear2 to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}
	birthYear1 := "20"
	if birthYear2Int >= 54 {
		birthYear1 = "19"
	}
	birthYear, err := strconv.Atoi(birthYear1 + birthYear2)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert birthYear to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	participantId, err := h.service.GetOrCreateParticipant(firstName, lastName, birthYear, userId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create or get participant")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	// check the course capacity before creating a new application
	capacityOK, err := h.service.CheckCourseCapacity(courseId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check course capacity")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}
	if !capacityOK {
		return utils.HTML(c, coursesTemplates.ApplicationFormErrorInfo("Kapacita kurzu se již bohužel vyčerpala. Zkuste prosím jiný kurz."))
	}

	// create a new application form
	_, err = h.service.CreateApplicationForm(courseId, participantId, personalId, parentName, phone, userEmail)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create application form")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	// send an email to the user and the admin
	err = h.service.SendApplicationFormEmail(userEmail, courseId, firstName, lastName, parentName, phone, birthYear1+birthYear2)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send application form email")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	successInfo := coursesTemplates.ApplicationFormSuccessInfo()
	return utils.HTML(c, successInfo)
}
