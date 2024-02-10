package security

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	loginTemplates "github.com/svachaj/sambar-wall/modules/security/templates"
	models "github.com/svachaj/sambar-wall/modules/security/types"
	"github.com/svachaj/sambar-wall/utils"
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

	loginModel := models.LoginFormModel{}
	loginModal := loginTemplates.LoginModal(loginModel)

	return utils.HTML(c, loginModal)
}

func (h *SecurityHandlers) SignIn(c echo.Context) error {

	// get user from db
	// check if user exists
	// if user exists, check if password is correct
	// if password is correct, set session and redirect to home
	// if password is incorrect, show error message
	// if user does not exist, show error message
	// username := c.FormValue("username")
	// password := c.FormValue("password")

	// var user types.User
	// err := h.db.Get(&user, "SELECT id, passwordhash, username FROM t_system_user tsu WHERE tsu.username = $1 or tsu.email = $1", username)

	// if err != nil {
	// 	return utils.HTMLWithStatus(c, loginTemplates.LoginError("Uživatel neexistuje"), 401)
	// }

	// authSession, _ := session.Get("auth", c)
	// authSession.Options = &sessions.Options{
	// 	Path:     "/",
	// 	MaxAge:   30, // 30 seconds
	// 	HttpOnly: true,
	// }

	//authSession.Save(c.Request(), c.Response())

	loginModel := models.LoginFormModel{}
	loginForm := loginTemplates.LoginForm(loginModel)

	return utils.HTML(c, loginForm)
}
