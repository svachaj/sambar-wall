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
	"github.com/svachaj/sambar-wall/modules/layouts"
	"github.com/svachaj/sambar-wall/modules/toasts"
	"github.com/svachaj/sambar-wall/utils"
)

type ICoursesHandler interface {
	GetCoursesList(c echo.Context) error
	ApplicationFormPage(c echo.Context) error
	ProcessApplicationForm(c echo.Context) error
	MyApplicationsPage(c echo.Context) error
	GetAllApplicationForms(c echo.Context) error
	SearchInApplications(c echo.Context) error
	SetApplicationFormPaid(c echo.Context) error
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

	isAuthenticated, _, _, roles := middlewares.IsAuthenticated(&c)

	coursesListComponent := coursesTemplates.CoursesList(courses, isAuthenticated)
	coursesPage := coursesTemplates.CoursesPage(coursesListComponent, isAuthenticated, middlewares.HasRole(roles, constants.ROLE_SAMBAR_ADMIN))

	return utils.HTML(c, coursesPage)
}

func (h *CoursesHandler) ApplicationFormPage(c echo.Context) error {

	id := c.Param("id")
	courseId, err := strconv.Atoi(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert courseId to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	// first check if the course is still available
	capacityOK, err := h.service.CheckCourseCapacity(courseId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check course capacity")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}
	if !capacityOK {
		return utils.HTML(c, layouts.BaseLayoutWithComponent(coursesTemplates.ApplicationFormErrorInfo("Kapacita kurzu se již bohužel vyčerpala. Zkuste prosím jiný kurz."), true, false))
	}

	courseInfo := h.service.GetCourseInfo(courseId)

	applicationForm := coursesTemplates.ApplicationFormPage(id, courseInfo)

	return utils.HTML(c, applicationForm)
}

func (h *CoursesHandler) ProcessApplicationForm(c echo.Context) error {

	// validate form
	params, _ := c.FormParams()
	courseIdString := params.Get(models.APPLICATION_FORM_COURSE_ID)
	courseId, err := strconv.Atoi(courseIdString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert courseId to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}
	applicationForm := models.ApplicationFormModel(courseIdString)

	isValid := applicationForm.ValidateFields(params)

	if !isValid {
		courseInfo := h.service.GetCourseInfo(courseId)
		applicationFormComponent := coursesTemplates.ApplicationForm(applicationForm, courseInfo, nil)
		return utils.HTML(c, applicationFormComponent)
	}

	// get values from the form
	firstName := applicationForm.FormFields[models.APPLICATION_FORM_FIRST_NAME].Value
	lastName := applicationForm.FormFields[models.APPLICATION_FORM_LAST_NAME].Value
	phone := applicationForm.FormFields[models.APPLICATION_FORM_PHONE].Value
	parentName := applicationForm.FormFields[models.APPLICATION_FORM_PARENT_NAME].Value
	healthState := applicationForm.FormFields[models.APPLICATION_FORM_HEALTH_STATE].Value

	personalId := applicationForm.FormFields[models.APPLICATION_FORM_PERSONAL_ID].Value

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
	birthYear2 := personalId[0:2]
	birthYear2Int, err := strconv.Atoi(birthYear2)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert birthYear2 to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}
	birthYear1 := "20"
	if len(personalId) == 9 {
		birthYear1 = "19"
	} else if birthYear2Int >= 54 {
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
	applicationFormId, err := h.service.CreateApplicationForm(courseId, participantId, personalId, parentName, phone, userEmail, userId, healthState)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create application form")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	// send an email to the user and the admin
	err = h.service.SendApplicationFormEmail(applicationFormId, userEmail, courseId, firstName, lastName, parentName, phone, birthYear1+birthYear2)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send application form email")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	successInfo := coursesTemplates.ApplicationFormSuccessInfo()
	return utils.HTML(c, successInfo)
}

func (h *CoursesHandler) MyApplicationsPage(c echo.Context) error {

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	userId := authSession.Values[constants.AUTH_USER_ID].(int)
	roles := authSession.Values[constants.AUTH_USER_ROLES].([]string)

	applications, err := h.service.GetApplicationsByUserId(userId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get applications by userId")
		return utils.HTML(c, httperrors.ErrorPage(httperrors.InternalServerErrorSimple()))
	}

	applicationsListComponent := coursesTemplates.MyApplicationsList(applications)
	applicationsPage := coursesTemplates.MyApplicationsPage(applicationsListComponent, middlewares.HasRole(roles, constants.ROLE_SAMBAR_ADMIN))

	return utils.HTML(c, applicationsPage)
}

func (h *CoursesHandler) GetAllApplicationForms(c echo.Context) error {

	searchText := c.QueryParam("searchText")
	applications, err := h.service.GetAllApplicationForms(searchText)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all applications")
		return utils.HTML(c, httperrors.ErrorPage(httperrors.InternalServerErrorSimple()))
	}

	applicationsListComponent := coursesTemplates.AllApplicationsList(applications)
	applicationsPage := coursesTemplates.AllApplicationsPage(applicationsListComponent)

	return utils.HTML(c, applicationsPage)
}

func (h *CoursesHandler) SearchInApplications(c echo.Context) error {

	searchText := c.QueryParam("search")
	applications, err := h.service.GetAllApplicationForms(searchText)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all applications")
		return utils.HTML(c, httperrors.ErrorPage(httperrors.InternalServerErrorSimple()))
	}

	if len(applications) == 0 {
		return utils.HTML(c, coursesTemplates.AllApplicationsNoApplications())
	} else {
		return utils.HTML(c, coursesTemplates.AllApplicationsTable(applications))
	}
}

func (h *CoursesHandler) SetApplicationFormPaid(c echo.Context) error {
	applicationFormIdParam := c.Param("id")
	applicationFormId, err := strconv.Atoi(applicationFormIdParam)
	if err != nil {
		log.Err(err).Msg("Can not parse application form id:" + applicationFormIdParam)
		return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(false, toasts.ErrorToast(constants.SOMETHING_GET_WRONG)))
	}

	paidParam := c.QueryParam("paid")
	paid, err := strconv.ParseBool(paidParam)

	if err != nil {
		log.Err(err).Msg("Can not parse paid param:" + paidParam)
		return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(false, toasts.ErrorToast(constants.SOMETHING_GET_WRONG)))
	}

	err = h.service.SetApplicationFormPaid(applicationFormId, paid)
	if err != nil {
		log.Err(err).Msg("Can not set paid on application form:" + applicationFormIdParam)
		return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(!paid, toasts.ErrorToast(constants.SOMETHING_GET_WRONG)))
	}

	return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(paid, toasts.SuccessToast(constants.SUCCESSFULLY_SET)))
}
