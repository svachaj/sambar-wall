package security

import (
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	db "github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/modules/constants"
	loginTemplates "github.com/svachaj/sambar-wall/modules/security/templates"
	types "github.com/svachaj/sambar-wall/modules/security/types"
	"github.com/svachaj/sambar-wall/utils"
	"golang.org/x/crypto/bcrypt"
)

type ISecurityHandlers interface {
	LoginModal(c echo.Context) error
	SignIn(c echo.Context) error
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
	err := h.db.Get(&user, "SELECT id, passwordhash, username FROM t_system_user tsu WHERE tsu.username = $1 or tsu.email = $1", username)

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
				MaxAge:   30, // 30 seconds
				HttpOnly: true,
			}

			authSession.Values[constants.AUTH_USER_KEY] = user.ID

			authSession.Save(c.Request(), c.Response())

			return c.Redirect(200, "/")
		}
	}

	loginForm := loginTemplates.LoginForm(loginModel)

	return utils.HTML(c, loginForm)
}
