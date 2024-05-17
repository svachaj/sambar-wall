package main

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/config"
	"github.com/svachaj/sambar-wall/db"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/modules/agreement"
	"github.com/svachaj/sambar-wall/modules/courses"
	httperrors "github.com/svachaj/sambar-wall/modules/http-errors"
	"github.com/svachaj/sambar-wall/modules/security"
	"github.com/svachaj/sambar-wall/utils"
)

func InitializeModulesAndMapRoutes(e *echo.Echo, settings *config.Config) error {

	db, err := db.Initialize(settings)

	if err != nil {
		return err
	}

	var emailService utils.IEmailService

	if settings.AppEnv == config.APP_ENV_LOCALHOST {
		emailService = utils.NewMockEmailService()
	} else {
		emailService = utils.NewEmailService(settings.SmtpHost, settings.SmtpPort, settings.SmtpUser, settings.SmtpPassword)
	}

	errorsHandlers := httperrors.NewErrorsHandler()
	httperrors.MapErrorsRoutes(e, errorsHandlers)
	log.Info().Msg("Module Errors Initialized and Routes Mapped Successfully.")

	agreementService := agreement.NewAgreementService(db, emailService)
	agreementHandlers := agreement.NewAgreementHandlers(agreementService)
	agreement.MapAgreementRoutes(e, agreementHandlers)
	log.Info().Msg("Module Agreement Initialized and Routes Mapped Successfully.")

	coursesService := courses.NewCoursesService(db, emailService)
	coursesHandlers := courses.NewCoursesHandler(coursesService)
	courses.MapCoursesRoutes(e, coursesHandlers)
	log.Info().Msg("Module Courses Initialized and Routes Mapped Successfully.")

	securityService := security.NewSecurityService(db, emailService)
	securityHandlers := security.NewSecurityHandlers(db, securityService, coursesService)
	security.MapSecurityRoutes(e, securityHandlers)
	log.Info().Msg("Module Security Initialized and Routes Mapped Successfully.")

	// validation handlers
	e.POST("/validate-form-field", middlewares.ValidateFormField)

	return nil

}
