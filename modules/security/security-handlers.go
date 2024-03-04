package security

import (
	"fmt"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	db "github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/modules/constants"
	"github.com/svachaj/sambar-wall/modules/home"
	loginTemplates "github.com/svachaj/sambar-wall/modules/security/templates"
	types "github.com/svachaj/sambar-wall/modules/security/types"
	"github.com/svachaj/sambar-wall/utils"
	"golang.org/x/crypto/bcrypt"
)

type ISecurityHandlers interface {
	LoginModal(c echo.Context) error
	SignIn(c echo.Context) error
	SignOut(c echo.Context) error
}

type SecurityHandlers struct {
	db *sqlx.DB
}

func NewSecurityHandlers(db *sqlx.DB) ISecurityHandlers {
	return &SecurityHandlers{db: db}
}

func (h *SecurityHandlers) LoginModal(c echo.Context) error {

	loginModel := types.LoginFormInitModel

	loginModal := loginTemplates.LoginModal(loginModel)

	return utils.HTML(c, loginModal)
}

func (h *SecurityHandlers) SignIn(c echo.Context) error {

	loginModel := types.LoginFormInitModel

	username := c.FormValue("username")
	password := c.FormValue("password")

	var user db.User
	query := fmt.Sprintf("SELECT id, passwordhash, username FROM t_system_user tsu WHERE lower(tsu.username) = '%[1]s' or tsu.email = '%[1]s' ", strings.ToLower(username))
	log.Info().Msg(query)
	err := h.db.Get(&user, query)

	if err != nil {
		loginModel.Errors = append(loginModel.Errors, types.ERROR_LOGIN)
		log.Err(err).Msg("Unathorized")
	} else {
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))

		if err != nil {
			loginModel.Errors = append(loginModel.Errors, types.ERROR_LOGIN)
			log.Err(err).Msg("Unathorized")
		} else {

			authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
			authSession.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   3600, // 3600 seconds
				HttpOnly: true,
			}

			authSession.Values[constants.AUTH_USER_KEY] = user.ID

			authSession.Save(c.Request(), c.Response())

			return utils.HTML(c, home.HomePage(h.db, true))
		}
	}

	loginForm := loginTemplates.LoginForm(loginModel)
	// there was an error, so we want to retarget to the login form again
	c.Response().Header().Set("HX-Retarget", "#login-form")

	return utils.HTML(c, loginForm)
}

func (h *SecurityHandlers) SignOut(c echo.Context) error {

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)
	authSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1, // delete cookie
		HttpOnly: true,
	}

	authSession.Save(c.Request(), c.Response())

	return utils.HTML(c, home.HomePage(h.db, false))
}
