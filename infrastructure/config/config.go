package config

// Layer: infrastructure (config)
// Peran: Memuat konfigurasi aplikasi dari environment variable (.env).

import (
    "fmt"
    "os"
    "strconv"
    "strings"

    "github.com/joho/godotenv"
)

// Config berisi konfigurasi dasar untuk koneksi InfluxDB.
type Config struct {
    InfluxURL   string
    InfluxToken string
    InfluxOrg   string
    InfluxBucket string
    // Security
    APIKey              string
    APISecret           string
    AllowedIPs          []string
    IPWhitelistEnabled  bool
    RateLimitPerMinute  int
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
        APIKey:       os.Getenv("API_KEY"),
        APISecret:    os.Getenv("API_SECRET"),
    }

    // Allowed IPs (comma separated)
    if ips := strings.TrimSpace(os.Getenv("ALLOWED_IPS")); ips != "" {
        parts := strings.Split(ips, ",")
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }
        cfg.AllowedIPs = parts
    } else {
        cfg.AllowedIPs = []string{}
    }

    // Toggle IP whitelist (default: disabled)
    switch strings.ToLower(strings.TrimSpace(os.Getenv("ENABLE_IP_WHITELIST"))) {
    case "1", "true", "yes", "on":
        cfg.IPWhitelistEnabled = true
    default:
        cfg.IPWhitelistEnabled = false
    }

    // Rate limit per minute (default: 30)
    rl := strings.TrimSpace(os.Getenv("RATE_LIMIT_PER_MINUTE"))
    if rl == "" {
        cfg.RateLimitPerMinute = 30
    } else {
        if v, err := strconv.Atoi(rl); err == nil && v > 0 {
            cfg.RateLimitPerMinute = v
        } else {
            cfg.RateLimitPerMinute = 30
        }
    }

    // Validasi minimal
    if cfg.InfluxURL == "" || cfg.InfluxToken == "" || cfg.InfluxOrg == "" || cfg.InfluxBucket == "" {
        return Config{}, fmt.Errorf("missing required env: INFLUX_URL/INFLUX_TOKEN/INFLUX_ORG/INFLUX_BUCKET")
    }
    return cfg, nil
}