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
    // Versi non-streaming untuk personel
    apiGroup.GET("/sensors/personel/list", s.GetPersonelList)
    // Endpoint Radar: sama seperti personel, bedanya measurement menggunakan "radar_sensor"
    apiGroup.POST("/sensors/radar", s.PostRadar)
    apiGroup.GET("/sensors/radar", s.GetRadar)
    // Versi non-streaming untuk radar
    apiGroup.GET("/sensors/radar/list", s.GetRadarList)
    // Endpoint DF: sama seperti personel, measurement menggunakan "df_sensor"
    apiGroup.POST("/sensors/df", s.PostDF)
    apiGroup.GET("/sensors/df", s.GetDF)
    // Versi non-streaming untuk DF
    apiGroup.GET("/sensors/df/list", s.GetDFList)
    // Endpoint ADSB: sama seperti personel, measurement menggunakan "adsb_sensor"
    apiGroup.POST("/sensors/adsb", s.PostADSB)
    apiGroup.GET("/sensors/adsb", s.GetADSB)
    // Versi non-streaming untuk ADSB
    apiGroup.GET("/sensors/adsb/list", s.GetADSBList)
    // Endpoint Available Dates untuk Personel
    apiGroup.GET("/sensors/available-data/personel", s.GetAvailableDatesPersonel)
    // Endpoint Available Dates untuk Radar
    apiGroup.GET("/sensors/available-data/radar", s.GetAvailableDatesRadar)
    // Endpoint Available Dates untuk ADSB
    apiGroup.GET("/sensors/available-data/adsb", s.GetAvailableDatesADSB)
    // Endpoint Available Dates untuk DF
    apiGroup.GET("/sensors/available-data/df", s.GetAvailableDatesDF)
    apiGroup.POST("/test/influx", s.TestStore)
}