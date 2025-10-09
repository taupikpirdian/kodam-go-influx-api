package main

// Layer: cmd
// Peran: Entry point aplikasi. Melakukan dependency injection sederhana,
// inisialisasi router, dan menjalankan HTTP server.

import (
    "context"
    "log"

    "github.com/labstack/echo/v4"
    "github.com/SALT-Indonesia/salt-pkg/logmanager"
    "github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmecho"

    "rti/influxdb/infrastructure/config"
    "rti/influxdb/infrastructure/repository"
    "rti/influxdb/infrastructure/router"
    handlerPkg "rti/influxdb/interface/http/handler"
    "rti/influxdb/usecase"
)

func main() {
    // Init framework HTTP Echo
    e := echo.New()

    // Integrasi Standard Logs (SALT LogManager)
    // Peran: Menyediakan structured logging untuk setiap request dengan middleware Echo.
    // Secara default, environment diatur otomatis berdasarkan APP_ENV.
    // Lihat: https://github.com/SALT-Indonesia/salt-pkg/tree/main/logmanager
    app := logmanager.NewApplication(
        logmanager.WithService("influxdb-api"),
    )
    e.Use(lmecho.Middleware(app))

    // Load configuration (.env)
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // Dependency Injection (manual/simple)
    // Health
    healthUC := usecase.NewHealthUsecase()
    healthHandler := handlerPkg.NewHealthHandler(healthUC)

    // Influx Repository
    influxRepo := repository.NewInfluxRepository(cfg)
    defer influxRepo.Close()

    // Cek koneksi ke InfluxDB saat startup
    if err := influxRepo.CheckConnection(context.Background()); err != nil {
        log.Fatalf("failed to connect to InfluxDB: %v", err)
    }

    // Sensor Usecase & Handler
    sensorUC := usecase.NewSensorUsecase(influxRepo)
    sensorHandler := handlerPkg.NewSensorHandler(sensorUC)

    // Registrasi rute dengan middleware keamanan berdasarkan config
    router.Register(e, healthHandler, sensorHandler, cfg)

    // Jalankan server
    if err := e.Start(":3000"); err != nil {
        log.Fatalf("server exited: %v", err)
    }
}