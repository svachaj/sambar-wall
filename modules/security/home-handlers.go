package security

import (
	"github.com/labstack/echo/v4"
	loginTemplates "github.com/svachaj/sambar-wall/modules/security/templates"
	"github.com/svachaj/sambar-wall/utils"
)

type ISecurityHandlers interface {
	LoginModal(c echo.Context) error
}

type SecurityHandlers struct {
}

func NewSecurityHandlers() ISecurityHandlers {
	return &SecurityHandlers{}
}

func (h *SecurityHandlers) LoginModal(c echo.Context) error {
	loginModal := loginTemplates.LoginModal()

	return utils.HTML(c, loginModal)
}
