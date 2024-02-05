package security

import (
	"github.com/labstack/echo/v4"
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
	loginModal := LoginModal()

	return utils.HTML(c, loginModal)
}
