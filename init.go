package main

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/modules/home"
	httperrors "github.com/svachaj/sambar-wall/modules/http-errors"
	"github.com/svachaj/sambar-wall/modules/security"
)

func InitializeModulesAndMapRoutes(e *echo.Echo) {

	securityHandlers := security.NewSecurityHandlers()
	security.MapSecurityRoutes(e, securityHandlers)
	log.Info().Msg("Module Security Initialized and Routes Mapped Successfully.")

	homeHandlers := home.NewHomeHandlers()
	home.MapHomeRoutes(e, homeHandlers)
	log.Info().Msg("Module Home Initialized and Routes Mapped Successfully.")

	errorsHandlers := httperrors.NewErrorsHandler()
	httperrors.MapErrorsRoutes(e, errorsHandlers)
	log.Info().Msg("Module Errors Initialized and Routes Mapped Successfully.")

}
