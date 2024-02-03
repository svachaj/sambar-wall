package main

import (
	"context"
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

func main() {
	// init logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("SAMBAR APP Starting...")

	// configuration settings
	// application enviroment varibles described in example.env file
	log.Info().Msg("Loading configuration...")
	settings, err := config.LoadConfiguraion()
	utils.PanicOnError(err)

	// Echo instance
	e := echo.New()

	// Echo Logging
	e.Use(middlewares.RequestLoggerWithConfig())

	// Recover from panics
	e.Use(middleware.Recover())

	// static files
	e.Static("/static", "static")

	// Initialize modules and map routes
	InitializeModulesAndMapRoutes(e)

	// Start server
	go func() {
		if err := e.Start(":" + settings.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server: SAMABR APP: ", err)
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
