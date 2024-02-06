package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/config"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	APP_NAME = "SAMBAR APP - COURSES AND REGISTRATION SYSTEM"
)

func main() {
	// init logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg(fmt.Sprintf("Starting %s", APP_NAME))

	// configuration settings
	// application enviroment varibles described in example.env file
	log.Info().Msg("Loading configuration...")
	settings, err := config.LoadConfiguraion()
	utils.PanicOnError(err)

	// Echo - http framewrok instance
	e := echo.New()

	// Echo Logging
	e.Use(middlewares.RequestLoggerWithConfig())

	// Recover from panics
	e.Use(middleware.Recover())

	// static files
	e.Static("/static", "static")

	// Initialize modules and map routes
	err = InitializeModulesAndMapRoutes(e, settings)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing modules and mapping routes.")
	}

	// Custom HTTP Error Handler to serve error pages
	e.HTTPErrorHandler = middlewares.CustomHTTPErrorHandler

	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", settings.AppPort)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server: ", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
