package config

// Layer: infrastructure (config)
// Peran: Memuat konfigurasi aplikasi dari environment variable (.env).

import (
    "fmt"
    "os"

    "github.com/joho/godotenv"
)

// Config berisi konfigurasi dasar untuk koneksi InfluxDB.
type Config struct {
    InfluxURL   string
    InfluxToken string
    InfluxOrg   string
    InfluxBucket string
}

// Load membaca konfigurasi dari environment variable.
// Secara default mencoba memuat file .env jika tersedia.
func Load() (Config, error) {
    // Muat file .env jika ada (abaikan error jika tidak ada)
    _ = godotenv.Load()

    cfg := Config{
        InfluxURL:    os.Getenv("INFLUX_URL"),
        InfluxToken:  os.Getenv("INFLUX_TOKEN"),
        InfluxOrg:    os.Getenv("INFLUX_ORG"),
        InfluxBucket: os.Getenv("INFLUX_BUCKET"),
    }

    // Validasi minimal
    if cfg.InfluxURL == "" || cfg.InfluxToken == "" || cfg.InfluxOrg == "" || cfg.InfluxBucket == "" {
        return Config{}, fmt.Errorf("missing required env: INFLUX_URL/INFLUX_TOKEN/INFLUX_ORG/INFLUX_BUCKET")
    }
    return cfg, nil
}