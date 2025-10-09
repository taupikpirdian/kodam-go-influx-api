package router

// Layer: infrastructure
// Peran: Detail implementasi teknis seperti konfigurasi router, middleware,
// storage, dsb. Berinteraksi dengan framework/driver.

import (
    "github.com/labstack/echo/v4"

    handlerPkg "rti/influxdb/interface/http/handler"
)

// Register mendaftarkan semua rute aplikasi pada Echo.
func Register(e *echo.Echo, h *handlerPkg.HealthHandler, s *handlerPkg.SensorHandler) {
    // Di sini Anda dapat menambahkan middleware global jika diperlukan.
    // e.Use(middleware.Logger())
    // e.Use(middleware.Recover())

    e.GET("/health", h.Health)
    e.POST("/api/sensors/personel", s.PostPersonel)
    e.GET("/api/sensors/personel", s.GetPersonel)
    e.POST("/api/test/influx", s.TestStore)
}