// SecurityHandlers is a struct that handles security-related operations such as login, sign-out, and user account management.
// It implements the ISecurityHandlers interface.
//
// Fields:
//   - db: *sqlx.DB - the database connection.
//   - coursesService: courses.ICoursesService - the service for handling course-related operations.
//   - securityService: ISecurityService - the service for handling security-related operations.
//
// Methods:
//   - Login(c echo.Context) error: Handles the login page rendering. If the "expired" query parameter is set to "true",
//     it indicates that the session has expired, and the login page will reflect that.
//   - SignOut(c echo.Context) error: Handles the sign-out process. It clears the authentication session and redirects
//     the user to the courses page.
//   - SignInStep1(c echo.Context) error: Handles the first step of the sign-in process. It validates the form input,
//     generates a verification code, saves it, and sends it to the user's email. If successful, it redirects to the second step of the sign-in process.
//   - SignInStep2(c echo.Context) error: Handles the second step of the sign-in process. It validates the form input,
//     finalizes the login process with the provided email and confirmation code, sets up the user session, and redirects the user to the appropriate page.
//   - SignMeIn(c echo.Context) error: Handles the sign-in process via a query parameter. It decodes the parameter,
//     finalizes the login process, sets up the user session, and redirects the user to the appropriate page.
//   - UserAccountPage(c echo.Context) error: Renders the user account page with the user's email.
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
	coursesPage := coursesTemplates.CoursesPage(coursesListComponent, false, false)

	return utils.HTML(c, coursesPage)
}

// SignInStep1 handles the first step of the sign-in process. It validates the form input,
// generates a verification code, saves it, and sends it to the user's email. If successful,
// it redirects to the second step of the sign-in process.
//
// Parameters:
//   - c: echo.Context - the context of the request, which provides request and response objects.
//
// Returns:
//   - error: an error if the sign-in process fails, otherwise nil.
//
// The function performs the following steps:
//  1. Logs the request path for debugging purposes.
//  2. Validates the form input using the SignInStep1InitModel.
//  3. If the form is invalid, it returns an HTML response with the form and validation errors.
//  4. Retrieves the email from the form fields.
//  5. Generates a verification code and saves it using the security service.
//  6. If saving the verification code fails, logs the error and returns an HTML response with a server error toast.
//  7. Sends the verification code to the user's email.
//  8. If sending the verification code fails, logs the error and returns an HTML response with a server error toast.
//  9. If everything is successful, initializes the second step form, pre-fills the email field, and returns an HTML response with the form and an info toast.
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

// SignInStep2 handles the second step of the sign-in process.
// It validates the form input, finalizes the login process, and sets up the user session.
//
// Parameters:
//   - c: echo.Context - the context of the request, which provides request and response objects.
//
// Returns:
//   - error: an error if the sign-in process fails, otherwise nil.
//
// The function performs the following steps:
//  1. Validates the form input using the SignInStep2InitModel.
//  2. If the form is invalid, it returns an HTML response with the form and validation errors.
//  3. Retrieves the email and confirmation code from the form fields.
//  4. Calls the security service to finalize the login process with the provided email and confirmation code.
//  5. If the login fails, logs the error, updates the form with an error message, and returns an HTML response with the form and error message.
//  6. Sets up the user session with the user's email, ID, and roles.
//  7. Retrieves the return URL from the session, if available, and sets the HX-Redirect header to redirect the user to the appropriate page.
//  8. Logs the successful sign-in event with the user's email, roles, and IP address.
//  9. Returns an HTML response with the form and a success message.
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
	email := strings.ToLower(step2Form.FormFields[models.LOGIN_FORM_EMAIL].Value)
	confirmationCode := step2Form.FormFields[models.LOGIN_FORM_CONFIRMATION_CODE].Value

	userId, roles, err := h.securityService.FinalizeLogin(email, confirmationCode)

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
	authSession.Values[constants.AUTH_USER_ROLES] = roles
	returnUrlInterf := authSession.Values[constants.AUTH_RETURN_URL]
	returnUrl := ""
	if returnUrlInterf != nil {
		returnUrl = returnUrlInterf.(string)
	}

	authSession.Save(c.Request(), c.Response())

	ipAddress := c.RealIP()
	log.Info().Msgf("User %v roles(%v) signed in from IP %v", email, roles, ipAddress)

	// if user is authenticated, we want to retarget to the courses page

	if returnUrl != "" {
		c.Response().Header().Set("HX-Redirect", returnUrl)
	} else {
		c.Response().Header().Set("HX-Redirect", constants.ROUTE_HOME)
	}

	step2 := security.LoginFormStep2(step2Form, toasts.SuccessToast("Přihlášení proběhlo úspěšně."))
	return utils.HTML(c, step2)
}

// SignMeIn handles the user sign-in process. It retrieves and decodes the query parameter,
// finalizes the login process, and sets up the authentication session.
//
// Parameters:
//   - c: echo.Context - the context for the request, which provides query parameters and request/response handling.
//
// Returns:
//   - error: an error if the sign-in process fails, otherwise nil.
//
// The function performs the following steps:
//  1. Retrieves the encoded query parameter "c" from the request and decodes it using the application secret.
//  2. Splits the decoded parameter to extract the email and confirmation code.
//  3. Calls the security service to finalize the login process with the extracted email and confirmation code.
//  4. If the login fails, logs an unauthorized error and redirects to the login route with an expired flag.
//  5. If the login succeeds, sets up the authentication session with user details and saves the session.
//  6. Logs the successful sign-in event with the user's email, roles, and IP address.
//  7. Redirects the user to the return URL if available, otherwise redirects to the courses page.
func (h *SecurityHandlers) SignMeIn(c echo.Context) error {

	// get query param and decode it
	queryEncodedParam := c.QueryParam("c")
	decodedParam := utils.Decrypt(queryEncodedParam, h.securityService.GetConfig().AppSecret)
	params := strings.Split(decodedParam, ";")

	email := strings.ToLower(params[0])
	confirmationCode := params[1]

	userId, roles, err := h.securityService.FinalizeLogin(email, confirmationCode)

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
	authSession.Values[constants.AUTH_USER_ROLES] = roles

	authSession.Save(c.Request(), c.Response())

	ipAddress := c.RealIP()
	log.Info().Msgf("User %v roles(%v) signed in from IP %v", email, roles, ipAddress)
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
