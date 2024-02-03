package main

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/modules/home"
)

func InitializeModulesAndMapRoutes(e *echo.Echo) {

	homeHandlers := home.NewHomeHandlers()
	home.MapHomeRoutes(e, homeHandlers)
	log.Info().Msg("Module Home Initialized and Routes Mapped Successfully.")

}
