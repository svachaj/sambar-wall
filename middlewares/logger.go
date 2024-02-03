package middlewares

import (
	"github.com/rs/zerolog/log"

	// echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RequestLoggerWithConfig() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogMethod:    true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.Error().
					Str("URI", v.URI).
					Int("status", v.Status).
					Str("latency", v.Latency.String()).
					Str("method", v.Method).
					Str("remote_ip", v.RemoteIP).
					Str("user_agent", v.UserAgent).
					Msg(v.Error.Error())
				return nil
			} else {
				if v.Status >= 500 {
					log.Error().
						Str("URI", v.URI).
						Int("status", v.Status).
						Str("latency", v.Latency.String()).
						Str("method", v.Method).
						Str("remote_ip", v.RemoteIP).
						Str("user_agent", v.UserAgent).
						Msg("")
					return nil
				} else if v.Status >= 400 {
					log.Warn().
						Str("URI", v.URI).
						Int("status", v.Status).
						Str("latency", v.Latency.String()).
						Str("method", v.Method).
						Str("remote_ip", v.RemoteIP).
						Str("user_agent", v.UserAgent).
						Msg("")
					return nil
				} else {
					log.Info().
						Str("URI", v.URI).
						Int("status", v.Status).
						Str("latency", v.Latency.String()).
						Str("method", v.Method).
						Str("remote_ip", v.RemoteIP).
						Str("user_agent", v.UserAgent).
						Msg("")
					return nil
				}
			}
		},
	})
}
