package router

// Layer: infrastructure
// Peran: Detail implementasi teknis seperti konfigurasi router, middleware,
// storage, dsb. Berinteraksi dengan framework/driver.

import (
    "github.com/labstack/echo/v4"

    handlerPkg "rti/influxdb/interface/http/handler"
    "rti/influxdb/infrastructure/config"
    secmw "rti/influxdb/infrastructure/middleware"
)

// Register mendaftarkan semua rute aplikasi pada Echo.
func Register(e *echo.Echo, h *handlerPkg.HealthHandler, s *handlerPkg.SensorHandler, cfg config.Config) {
    // Di sini Anda dapat menambahkan middleware global jika diperlukan.
    // e.Use(middleware.Logger())
    // e.Use(middleware.Recover())

    // Apply CORS globally if enabled
    cors := secmw.CORSMiddleware(cfg)
    for _, m := range cors {
        e.Use(m)
    }

    e.GET("/health", h.Health)

    // Group /api dengan middleware keamanan
    apiGroup := e.Group("/api", secmw.SecurityMiddleware(cfg)...) 

    apiGroup.POST("/sensors/personel", s.PostPersonel)
    apiGroup.GET("/sensors/personel", s.GetPersonel)
    // Endpoint Radar: sama seperti personel, bedanya measurement menggunakan "radar_sensor"
    apiGroup.POST("/sensors/radar", s.PostRadar)
    apiGroup.GET("/sensors/radar", s.GetRadar)
    // Endpoint DF: sama seperti personel, measurement menggunakan "df_sensor"
    apiGroup.POST("/sensors/df", s.PostDF)
    apiGroup.GET("/sensors/df", s.GetDF)
    // Endpoint ADSB: sama seperti personel, measurement menggunakan "adsb_sensor"
    apiGroup.POST("/sensors/adsb", s.PostADSB)
    apiGroup.GET("/sensors/adsb", s.GetADSB)
    apiGroup.POST("/test/influx", s.TestStore)
}