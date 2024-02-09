package security

import (
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	loginTemplates "github.com/svachaj/sambar-wall/modules/security/templates"
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

	authSession, _ := session.Get("auth", c)

	signInCounts := 0

	if authSession.Values["numberOfSignIns"] != nil {
		signInCounts = authSession.Values["numberOfSignIns"].(int)
	}

	loginModal := loginTemplates.LoginModal(signInCounts)

	return utils.HTML(c, loginModal)
}

func (h *SecurityHandlers) SignIn(c echo.Context) error {

	authSession, _ := session.Get("auth", c)
	authSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   30,
		HttpOnly: true,
	}

	signInCounts := 0
	signIns := authSession.Values["numberOfSignIns"]

	if signIns == nil {
		signInCounts = 1
	} else {
		signInCounts = signIns.(int) + 1
	}

	authSession.Values["numberOfSignIns"] = signInCounts

	authSession.Save(c.Request(), c.Response())

	loginForm := loginTemplates.SignInsInfo(signInCounts)

	return utils.HTML(c, loginForm)
}
