package home

import (
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/utils"
)

type IHomeHandlers interface {
	Home(c echo.Context) error
}

type HomeHandlers struct {
}

func NewHomeHandlers() IHomeHandlers {
	return &HomeHandlers{}
}

func (h *HomeHandlers) Home(c echo.Context) error {
	homeIndex := HomeIndex()
	homeComponent := Home(homeIndex)

	return utils.HTML(c, homeComponent)
}
