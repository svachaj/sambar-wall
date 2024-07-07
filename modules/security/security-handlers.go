package security

import (
	"fmt"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/modules/constants"
	"github.com/svachaj/sambar-wall/modules/courses"
	coursesTemplates "github.com/svachaj/sambar-wall/modules/courses/templates"
	"github.com/svachaj/sambar-wall/modules/security/models"
	security "github.com/svachaj/sambar-wall/modules/security/templates"
	types "github.com/svachaj/sambar-wall/modules/security/types"
	"github.com/svachaj/sambar-wall/modules/toasts"
	"github.com/svachaj/sambar-wall/utils"
)

const (
	EXPIRATION_SESSION_TIME_SECONDS = 60 * 60 * 48 // 48 hours
)

type ISecurityHandlers interface {
	Login(c echo.Context) error
	SignInStep1(c echo.Context) error
	SignInStep2(c echo.Context) error
	SignOut(c echo.Context) error
	UserAccountPage(c echo.Context) error
	SignMeIn(c echo.Context) error
}

type SecurityHandlers struct {
	db              *sqlx.DB
	coursesService  courses.ICoursesService
	securityService ISecurityService
}

func NewSecurityHandlers(db *sqlx.DB, securityService ISecurityService, coursesService courses.ICoursesService) ISecurityHandlers {
	return &SecurityHandlers{db: db, coursesService: coursesService, securityService: securityService}
}

func (h *SecurityHandlers) Login(c echo.Context) error {

	expired := c.QueryParam("expired")

	loginPage := security.LoginPage(expired == "true")

	return utils.HTML(c, loginPage)
}

func (h *SecurityHandlers) SignOut(c echo.Context) error {

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	authSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1, // delete cookie
		HttpOnly: true,
	}

	authSession.Save(c.Request(), c.Response())

	courses, err := h.coursesService.GetCoursesList()

	if err != nil {
		return c.String(500, "Internal Server Error")
	}

	coursesListComponent := coursesTemplates.CoursesList(courses, false)
	coursesPage := coursesTemplates.CoursesPage(coursesListComponent, false)

	return utils.HTML(c, coursesPage)
}

func (h *SecurityHandlers) SignInStep1(c echo.Context) error {

	log.Info().Msg(c.Path())

	// validate form
	step1Form := models.SignInStep1InitModel()
	params, _ := c.FormParams()
	isFormValid := step1Form.ValidateFields(params)

	if !isFormValid {
		step1 := security.LoginFormStep1(step1Form, nil)
		return utils.HTML(c, step1)
	}

	// get specific form fields from the form
	email := step1Form.FormFields[models.LOGIN_FORM_EMAIL].Value

	// generate and save sign-in code
	code := h.securityService.GenerateVerificationCode()
	err := h.securityService.SaveVerificationCode(email, code)
	if err != nil {
		log.Error().Msgf("Save verification code error: %v", err)
		step1WithToast := security.LoginFormStep1(step1Form, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	// send sign-in code to the user
	err = h.securityService.SendVerificationCode(email, code, c.Request().Header.Get("Origin"))
	if err != nil {
		log.Error().Msgf("Send verification code error: %v", err)
		step1WithToast := security.LoginFormStep1(step1Form, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	// if everything is ok, we want to retarget to the step 2 of the sign-in process
	step2Form := models.SignInStep2InitModel()
	if val, ok := step2Form.FormFields[models.LOGIN_FORM_EMAIL]; ok {
		val.Value = email
		step2Form.FormFields[models.LOGIN_FORM_EMAIL] = val
	}
	step2 := security.LoginFormStep2(step2Form, toasts.InfoToast(fmt.Sprintf("Na zadaný email %v byl odeslán ověřovací kód.", email)))
	return utils.HTML(c, step2)
}

func (h *SecurityHandlers) SignInStep2(c echo.Context) error {

	// validate form
	step2Form := models.SignInStep2InitModel()
	params, _ := c.FormParams()
	isFormValid := step2Form.ValidateFields(params)

	if !isFormValid {
		step1 := security.LoginFormStep2(step2Form, nil)
		return utils.HTML(c, step1)
	}

	// get specific form fields from the form
	email := step2Form.FormFields[models.LOGIN_FORM_EMAIL].Value
	confirmationCode := step2Form.FormFields[models.LOGIN_FORM_CONFIRMATION_CODE].Value

	userId, err := h.securityService.FinalizeLogin(email, confirmationCode)

	if err != nil {
		log.Err(fmt.Errorf("Unathorized")).Msg("Unathorized")
		step2Form.Errors = append(step2Form.Errors, types.ERROR_LOGIN)
		step2 := security.LoginFormStep2(step2Form, toasts.ErrorToast(types.ERROR_LOGIN))
		return utils.HTML(c, step2)
	}

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	authSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   EXPIRATION_SESSION_TIME_SECONDS,
		HttpOnly: true,
	}

	authSession.Values[constants.AUTH_USER_USERNAME] = email
	authSession.Values[constants.AUTH_USER_ID] = userId
	returnUrlInterf := authSession.Values[constants.AUTH_RETURN_URL]
	returnUrl := ""
	if returnUrlInterf != nil {
		returnUrl = returnUrlInterf.(string)
	}

	authSession.Save(c.Request(), c.Response())

	ipAddress := c.RealIP()
	log.Info().Msgf("User %v signed in from IP %v", email, ipAddress)

	// if user is authenticated, we want to retarget to the courses page

	if returnUrl != "" {
		c.Response().Header().Set("HX-Redirect", returnUrl)
	} else {
		c.Response().Header().Set("HX-Redirect", constants.ROUTE_HOME)
	}

	step2 := security.LoginFormStep2(step2Form, toasts.SuccessToast("Přihlášení proběhlo úspěšně."))
	return utils.HTML(c, step2)
}

func (h *SecurityHandlers) SignMeIn(c echo.Context) error {

	// get query param and decode it
	queryEncodedParam := c.QueryParam("c")
	decodedParam := utils.Decrypt(queryEncodedParam)
	params := strings.Split(decodedParam, ";")

	email := params[0]
	confirmationCode := params[1]

	userId, err := h.securityService.FinalizeLogin(email, confirmationCode)

	if err != nil {
		log.Err(fmt.Errorf("Unathorized")).Msg("Unathorized")
		return c.Redirect(302, constants.ROUTE_LOGIN+"?expired=true")
	}

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	authSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   EXPIRATION_SESSION_TIME_SECONDS,
		HttpOnly: true,
	}

	returnUrlInterf := authSession.Values[constants.AUTH_RETURN_URL]
	returnUrl := ""
	if returnUrlInterf != nil {
		returnUrl = returnUrlInterf.(string)
	}

	authSession.Values[constants.AUTH_USER_USERNAME] = email
	authSession.Values[constants.AUTH_USER_ID] = userId

	authSession.Save(c.Request(), c.Response())

	ipAddress := c.RealIP()
	log.Info().Msgf("User %v signed in from IP %v", email, ipAddress)
	// if user is authenticated, we want to retarget to the courses page

	if returnUrl != "" {
		return c.Redirect(302, returnUrl)
	}

	return c.Redirect(302, constants.ROUTE_COURSES)
}

func (h *SecurityHandlers) UserAccountPage(c echo.Context) error {

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	userEmail := authSession.Values[constants.AUTH_USER_USERNAME].(string)

	userAccountPage := security.UserAccountPage(userEmail)

	return utils.HTML(c, userAccountPage)
}
