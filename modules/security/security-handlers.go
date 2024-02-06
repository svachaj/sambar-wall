package security

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	loginTemplates "github.com/svachaj/sambar-wall/modules/security/templates"
	"github.com/svachaj/sambar-wall/utils"
)

type ISecurityHandlers interface {
	LoginModal(c echo.Context) error
}

type SecurityHandlers struct {
	db *sqlx.DB
}

func NewSecurityHandlers(db *sqlx.DB) ISecurityHandlers {
	return &SecurityHandlers{db: db}
}

func (h *SecurityHandlers) LoginModal(c echo.Context) error {
	loginModal := loginTemplates.LoginModal()

	return utils.HTML(c, loginModal)
}
