package courses

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/constants"
	"github.com/svachaj/sambar-wall/modules/courses/models"
	coursesTemplates "github.com/svachaj/sambar-wall/modules/courses/templates"
	httperrors "github.com/svachaj/sambar-wall/modules/http-errors"
	"github.com/svachaj/sambar-wall/modules/layouts"
	"github.com/svachaj/sambar-wall/modules/toasts"
	"github.com/svachaj/sambar-wall/utils"

	"github.com/xuri/excelize/v2"
)

type ICoursesHandler interface {
	GetCoursesList(c echo.Context) error
	ApplicationFormPage(c echo.Context) error
	ProcessApplicationForm(c echo.Context) error
	MyApplicationsPage(c echo.Context) error
	GetAllApplicationForms(c echo.Context) error
	SearchInApplications(c echo.Context) error
	SetApplicationFormPaid(c echo.Context) error
	GetApplicationFormEditPage(c echo.Context) error
	UpdateApplicationForm(c echo.Context) error
	CancelApplicationFormEdit(c echo.Context) error
	BulkApplicationFormCreateWillContinue(c echo.Context) error
	CoursesAdminPage(c echo.Context) error
	ExportApplicationForms(c echo.Context) error
	ExportApplicationFormsInit(c echo.Context) error
	ExportApplicationFormsExcel(c echo.Context) error
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

	searchText := c.QueryParam("search")
	applications, err := h.service.GetAllApplicationForms(searchText)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all applications")
		return utils.HTML(c, httperrors.ErrorPage(httperrors.InternalServerErrorSimple()))
	}

	applicationsListComponent := coursesTemplates.AllApplicationsList(applications, searchText, nil)
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

	// add htmx push url to the search form
	if searchText == "" {
		c.Response().Header().Set("HX-Push-Url", "/prihlasky")
	} else {
		c.Response().Header().Set("HX-Push-Url", "/prihlasky?search="+searchText)
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
		return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(false, applicationFormIdParam, toasts.ErrorToast(constants.SOMETHING_GET_WRONG)))
	}

	paidParam := c.QueryParam("paid")
	paid, err := strconv.ParseBool(paidParam)

	if err != nil {
		log.Err(err).Msg("Can not parse paid param:" + paidParam)
		return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(false, applicationFormIdParam, toasts.ErrorToast(constants.SOMETHING_GET_WRONG)))
	}

	err = h.service.SetApplicationFormPaid(applicationFormId, paid)
	if err != nil {
		log.Err(err).Msg("Can not set paid on application form:" + applicationFormIdParam)
		return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(!paid, applicationFormIdParam, toasts.ErrorToast(constants.SOMETHING_GET_WRONG)))
	}

	return utils.HTML(c, coursesTemplates.ApplicationPaidInfoWithToast(paid, applicationFormIdParam, toasts.SuccessToast(constants.SUCCESSFULLY_SET)))
}

func (h *CoursesHandler) GetApplicationFormEditPage(c echo.Context) error {

	backUrl := c.Request().Referer()
	if backUrl == "" {
		backUrl = "/prihlasky"
	}

	applicationFormIdParam := c.Param("id")
	applicationFormId, err := strconv.Atoi(applicationFormIdParam)
	if err != nil {
		log.Err(err).Msg("Can not parse application form id:" + applicationFormIdParam)
		return utils.HTML(c, httperrors.ErrorPage(httperrors.InternalServerErrorSimple()))
	}

	applicationForm, err := h.service.GetApplicationFormById(applicationFormId)
	if err != nil {
		log.Err(err).Msg("Can not get application form by id:" + applicationFormIdParam)
		return utils.HTML(c, httperrors.ErrorPage(httperrors.InternalServerErrorSimple()))
	}

	applicationFormEditPage := coursesTemplates.ApplicationFormEditPage(applicationForm, backUrl)
	return utils.HTML(c, applicationFormEditPage)
}

func (h *CoursesHandler) UpdateApplicationForm(c echo.Context) error {

	// validate form
	params, _ := c.FormParams()
	applicationFormIdString := params.Get(models.APPLICATION_FORM_ID)
	applicationFormId, err := strconv.Atoi(applicationFormIdString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert applicationFormId to int")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	applicationForm := types.ApplicationForm{ID: applicationFormId}
	applicationFormModel := models.ApplicationFormEditModel(applicationForm)

	isValid := applicationFormModel.ValidateFields(params)

	if !isValid {
		applicationFormEditPage := coursesTemplates.ApplicationFormEditPage(applicationForm, "")
		return utils.HTML(c, applicationFormEditPage)
	}

	// get values from the form
	firstName := applicationFormModel.FormFields[models.APPLICATION_FORM_FIRST_NAME].Value
	lastName := applicationFormModel.FormFields[models.APPLICATION_FORM_LAST_NAME].Value
	phone := applicationFormModel.FormFields[models.APPLICATION_FORM_PHONE].Value
	parentName := applicationFormModel.FormFields[models.APPLICATION_FORM_PARENT_NAME].Value
	healthState := applicationFormModel.FormFields[models.APPLICATION_FORM_HEALTH_STATE].Value
	personalId := applicationFormModel.FormFields[models.APPLICATION_FORM_PERSONAL_ID].Value
	paid := applicationFormModel.FormFields[models.APPLICATION_FORM_PAID].Value
	isActive := applicationFormModel.FormFields[models.APPLICATION_FORM_IS_ACTIVE].Value

	err = h.service.UpdateApplicationForm(applicationFormId, personalId, parentName, healthState, firstName, lastName, phone, paid == "on", isActive == "on")

	if err != nil {
		log.Error().Err(err).Msg("Failed to update application form")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	_, searchParam, _ := utils.SetBackUrlAndGetQueryParamFromUrl(c, "search", "/prihlasky")

	applicationForms, err := h.service.GetAllApplicationForms(searchParam)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get all application forms")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	return utils.HTML(c, coursesTemplates.AllApplicationsList(applicationForms, searchParam, toasts.SuccessToast(constants.SUCCESSFULLY_UPDATED)))
}

func (h *CoursesHandler) CancelApplicationFormEdit(c echo.Context) error {

	_, searchParam, _ := utils.SetBackUrlAndGetQueryParamFromUrl(c, "search", "/prihlasky")

	applicationForms, err := h.service.GetAllApplicationForms(searchParam)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get all application forms")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	return utils.HTML(c, coursesTemplates.AllApplicationsList(applicationForms, searchParam, nil))
}

func (h *CoursesHandler) BulkApplicationFormCreateWillContinue(c echo.Context) error {

	// get all application forms that will continue
	applicationForms, err := h.service.GetApplicationFormsWillContinue()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get application forms that will continue")
		return echo.ErrInternalServerError
	}

	succesCount, errorCount := 0, 0
	// create a new application form for each application form that will continue
	for _, applicationForm := range applicationForms {
		if applicationForm.ParticipantID != nil && applicationForm.PersonalID != nil && applicationForm.ParentName != nil && applicationForm.Phone != nil && applicationForm.Email != nil && applicationForm.HealthState != nil && applicationForm.BirthYear != nil {
			newAppId, err := h.service.CreateApplicationForm(applicationForm.CourseID, *applicationForm.ParticipantID, *applicationForm.PersonalID, *applicationForm.ParentName, *applicationForm.Phone, *applicationForm.Email, applicationForm.CreatedByID, *applicationForm.HealthState)

			if err != nil {
				log.Error().Err(err).Msg("Failed to create application form")
				errorCount++
			} else {
				// send an email to the user and the admin
				err = h.service.SendApplicationFormEmail(newAppId, *applicationForm.Email, applicationForm.CourseID, applicationForm.FirstName, applicationForm.LastName, *applicationForm.ParentName, *applicationForm.Phone, strconv.Itoa(*applicationForm.BirthYear))

				if err != nil {
					log.Error().Err(err).Msg("Failed to send application form email")
					errorCount++
				} else {
					succesCount++
				}
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": succesCount,
		"error":   errorCount,
	})

}

func (h *CoursesHandler) CoursesAdminPage(c echo.Context) error {

	coursesPage := coursesTemplates.ApplicationsAdminPage()

	return utils.HTML(c, coursesPage)
}

func (h *CoursesHandler) ExportApplicationFormsInit(c echo.Context) error {
	html := `
	
	<script>document.getElementById('download-form').submit();</script>
	`

	time.Sleep(600 * time.Millisecond) // for fun :)

	return c.HTML(http.StatusOK, html)
}

func (h *CoursesHandler) ExportApplicationForms(c echo.Context) error {

	applicationForms, err := h.service.GetAllApplicationForms("")

	if err != nil {
		log.Error().Err(err).Msg("Failed to get all application forms")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=prihlasky-vse.csv")
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")

	// Convert applicationForms to [][]string
	csvData := make([][]string, len(applicationForms)+1)
	csvData[0] = []string{
		"ID",
		"Jméno",
		"Příjmení",
		"Rodné číslo",
		"Rok narození",
		"Telefon",
		"Název kurzu",
		"Cena kurzu",
		"Dny kurzu",
		"Čas od",
		"Čas do",
		"E-mail",
		"Jméno rodiče",
		"Zaplaceno",
		"Zdravotní stav",
		"Kód kurzu",
		"Věková skupina",
	} // Add headers

	for i, form := range applicationForms {
		personalID := ""
		phone := ""
		email := ""
		parentName := ""
		healthState := ""
		birthYer := ""
		if form.PersonalID != nil {
			personalID = *form.PersonalID
		}
		if form.Phone != nil {
			phone = *form.Phone
		}
		if form.Email != nil {
			email = *form.Email
		}
		if form.ParentName != nil {
			parentName = *form.ParentName
		}
		if form.HealthState != nil {
			healthState = *form.HealthState
		}
		if form.BirthYear != nil {
			birthYer = strconv.Itoa(*form.BirthYear)
		}
		// Add each application form to the CSV data
		csvData[i+1] = []string{
			strconv.Itoa(form.ID),
			form.FirstName,
			form.LastName,
			personalID,
			birthYer,
			phone,
			form.CourseName,
			strconv.FormatFloat(form.CoursePrice, 'f', 2, 64),
			form.CourseDays,
			form.CourseTimeFrom.Format("15:04"), // export only time in 24-hour format
			form.CourseTimeTo.Format("15:04"),
			email,
			parentName,
			strconv.FormatBool(form.Paid),
			healthState,
			form.CourseCode,
			form.CourseAgeGroup,
		}
	}

	err = utils.WriteCSV(c.Response(), csvData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write csv")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	return nil
}

func (h *CoursesHandler) ExportApplicationFormsExcel(c echo.Context) error {
	applicationForms, err := h.service.GetAllApplicationForms("")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all application forms")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	f := excelize.NewFile()
	sheet := "Přihlášky"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"ID", "Jméno", "Příjmení", "Rodné číslo", "Rok narození",
		"Telefon", "Název kurzu", "Cena kurzu", "Dny kurzu", "Čas od", "Čas do",
		"E-mail", "Jméno rodiče", "Zaplaceno", "Zdravotní stav",
		"Kód kurzu", "Věková skupina",
	}

	// Zapsání hlavičky
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Zapsání řádků s daty
	for rowIdx, form := range applicationForms {
		r := rowIdx + 2
		personalID := getString(form.PersonalID)
		phone := getString(form.Phone)
		email := getString(form.Email)
		parentName := getString(form.ParentName)
		healthState := getString(form.HealthState)
		birthYear := ""
		if form.BirthYear != nil {
			birthYear = strconv.Itoa(*form.BirthYear)
		}

		row := []interface{}{
			form.ID,
			form.FirstName,
			form.LastName,
			personalID,
			birthYear,
			phone,
			form.CourseName,
			utils.FormatPrice(form.CoursePrice),
			form.CourseDays,
			form.CourseTimeFrom.Format("15:04"),
			form.CourseTimeTo.Format("15:04"),
			email,
			parentName,
			form.Paid,
			healthState,
			form.CourseCode,
			form.CourseAgeGroup,
		}

		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, r)
			f.SetCellValue(sheet, cell, val)
		}
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=prihlasky-vse.xlsx")
	c.Response().Header().Set(echo.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	if err := f.Write(c.Response()); err != nil {
		log.Error().Err(err).Msg("Failed to write xlsx")
		return utils.HTML(c, httperrors.InternalServerErrorSimple())
	}

	return nil
}

func getString(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}
