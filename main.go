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
)

func main() {
	// init logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Sambar App Starting...")

	// configuration settings
	// application enviroment varibles described in example.env file
	log.Info().Msg("Loading configuration...")
	settings, err := config.LoadConfiguraion()
	utils.PanicOnError(err)

	// Echo instance
	e := echo.New()

	// Echo Logging
	e.Use(middlewares.RequestLoggerWithConfig())

	// static files
	e.Static("/", "static")

	// Start server
	go func() {
		if err := e.Start(":" + settings.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server: ELI - PANDA - API Gateway: ", err)
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
