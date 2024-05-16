package security

import (
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/modules/constants"
	"github.com/svachaj/sambar-wall/modules/courses"
	coursesTemplates "github.com/svachaj/sambar-wall/modules/courses/templates"
	"github.com/svachaj/sambar-wall/modules/security/models"
	loginTemplates "github.com/svachaj/sambar-wall/modules/security/templates"
	security "github.com/svachaj/sambar-wall/modules/security/templates"
	types "github.com/svachaj/sambar-wall/modules/security/types"
	"github.com/svachaj/sambar-wall/modules/toasts"
	"github.com/svachaj/sambar-wall/utils"
)

type ISecurityHandlers interface {
	Login(c echo.Context) error
	SignInStep1(c echo.Context) error
	SignInStep2(c echo.Context) error
	SignOut(c echo.Context) error
	UserAccountPage(c echo.Context) error
}

type SecurityHandlers struct {
	db             *sqlx.DB
	coursesService courses.ICoursesService
}

func NewSecurityHandlers(db *sqlx.DB, coursesService courses.ICoursesService) ISecurityHandlers {
	return &SecurityHandlers{db: db, coursesService: coursesService}
}

func (h *SecurityHandlers) Login(c echo.Context) error {

	loginPage := loginTemplates.LoginPage()

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

	// validate form
	step1Form := models.SignInStep1InitModel()
	params, _ := c.FormParams()
	isValid := step1Form.ValidateFields(params)

	if !isValid {
		step1 := security.LoginFormStep1(step1Form, nil)
		return utils.HTML(c, step1)
	}

	email := step1Form.FormFields[models.LOGIN_FORM_EMAIL].Value

	// generate and save verification code
	// code := h.service.GenerateVerificationCode()
	// err := h.service.SaveVerificationCode(email, code)
	// if err != nil {
	// 	log.Error().Msgf("Save verification code error: %v", err)
	// 	step1WithToast := agreementTemplates.Step1Form(step1Form, toasts.ServerErrorToast())
	// 	return utils.HTML(c, step1WithToast)
	// }

	// // send verification code
	// err = h.service.SendVerificationCode(email, code)
	// if err != nil {
	// 	log.Error().Msgf("Send verification code error: %v", err)
	// 	step1WithToast := agreementTemplates.Step1Form(step1Form, toasts.ServerErrorToast())
	// 	return utils.HTML(c, step1WithToast)
	// }

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
	isValid := step2Form.ValidateFields(params)

	if !isValid {
		step1 := security.LoginFormStep2(step2Form, nil)
		return utils.HTML(c, step1)
	}

	email := step2Form.FormFields[models.LOGIN_FORM_EMAIL].Value
	confirmationCode := step2Form.FormFields[models.LOGIN_FORM_CONFIRMATION_CODE].Value

	if confirmationCode != "1234" && email != "" {
		log.Err(fmt.Errorf("Unathorized")).Msg("Unathorized")
		step2Form.Errors = append(step2Form.Errors, types.ERROR_LOGIN)
		step2 := security.LoginFormStep2(step2Form, nil)
		return utils.HTML(c, step2)
	}

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	authSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 3600 seconds
		HttpOnly: true,
	}

	authSession.Values[constants.AUTH_USER_KEY] = email

	authSession.Save(c.Request(), c.Response())

	// if user is authenticated, we want to retarget to the courses page

	courses, err := h.coursesService.GetCoursesList()

	if err != nil {
		return c.String(500, "Internal Server Error")
	}

	coursesListComponent := coursesTemplates.CoursesList(courses, true)
	coursesPage := coursesTemplates.CoursesPage(coursesListComponent, true)

	c.Response().Header().Set("HX-Retarget", "body")
	return utils.HTML(c, coursesPage)

	// var user db.User
	// query := fmt.Sprintf("SELECT id, passwordhash, username FROM t_system_user tsu WHERE lower(tsu.username) = '%[1]s' or tsu.email = '%[1]s' ", strings.ToLower(username))
	// log.Info().Msg(query)
	// err := h.db.Get(&user, query)

	// if err != nil {
	// 	loginModel.Errors = append(loginModel.Errors, types.ERROR_LOGIN)
	// 	log.Err(err).Msg("Unathorized")
	// } else {
	// 	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))

	// 	if err != nil {
	// 		loginModel.Errors = append(loginModel.Errors, types.ERROR_LOGIN)
	// 		log.Err(err).Msg("Unathorized")
	// 	} else {

	// 		authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	// 		authSession.Options = &sessions.Options{
	// 			Path:     "/",
	// 			MaxAge:   3600, // 3600 seconds
	// 			HttpOnly: true,
	// 		}

	// 		authSession.Values[constants.AUTH_USER_KEY] = user.ID

	// 		authSession.Save(c.Request(), c.Response())

	// 		// if user is authenticated, we want to retarget to the courses page

	// 		courses, err := h.coursesService.GetCoursesList()

	// 		if err != nil {
	// 			return c.String(500, "Internal Server Error")
	// 		}

	// 		coursesListComponent := coursesTemplates.CoursesList(courses, true)
	// 		coursesPage := coursesTemplates.CoursesPage(coursesListComponent, true)

	// 		return utils.HTML(c, coursesPage)
	// 	}
	// }

	// loginForm := loginTemplates.LoginForm(loginModel)
	// // there was an error, so we want to retarget to the login form again
	// c.Response().Header().Set("HX-Retarget", "#login-form")

	// return utils.HTML(c, loginForm)
}

func (h *SecurityHandlers) UserAccountPage(c echo.Context) error {

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	userEmail := authSession.Values[constants.AUTH_USER_KEY].(string)

	userAccountPage := loginTemplates.UserAccountPage(userEmail)

	return utils.HTML(c, userAccountPage)
}
