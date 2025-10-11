package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"rti/influxdb/infrastructure/config"
)

// CORSMiddleware returns Echo CORS middleware configured from app config.
// If CORS is disabled, returns empty slice (no middleware applied).
func CORSMiddleware(cfg config.Config) []echo.MiddlewareFunc {
	if !cfg.CORSEnabled {
		return []echo.MiddlewareFunc{}
	}

	// If any origin contains wildcard '*', prefer AllowOriginFunc that allows all.
	allowAll := false
	for _, o := range cfg.CORSAllowedOrigins {
		if strings.TrimSpace(o) == "*" {
			allowAll = true
			break
		}
	}

	corsCfg := middleware.CORSConfig{
		AllowMethods:     cfg.CORSAllowedMethods,
		AllowHeaders:     cfg.CORSAllowedHeaders,
		ExposeHeaders:    cfg.CORSExposedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           cfg.CORSMaxAge,
	}
	if allowAll {
		corsCfg.AllowOriginFunc = func(origin string) (bool, error) { return true, nil }
	} else {
		corsCfg.AllowOrigins = cfg.CORSAllowedOrigins
	}

	return []echo.MiddlewareFunc{middleware.CORSWithConfig(corsCfg)}
}
