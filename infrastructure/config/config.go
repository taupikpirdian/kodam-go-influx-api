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
    // CORS
    CORSEnabled         bool
    CORSAllowedOrigins  []string
    CORSAllowedMethods  []string
    CORSAllowedHeaders  []string
    CORSExposedHeaders  []string
    CORSAllowCredentials bool
    CORSMaxAge          int
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

    // CORS configuration
    switch strings.ToLower(strings.TrimSpace(os.Getenv("CORS_ENABLED"))) {
    case "1", "true", "yes", "on":
        cfg.CORSEnabled = true
    default:
        cfg.CORSEnabled = false
    }

    // Allowed origins (comma separated). Use * to allow any.
    if s := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); s != "" {
        parts := strings.Split(s, ",")
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }
        cfg.CORSAllowedOrigins = parts
    } else {
        cfg.CORSAllowedOrigins = []string{"*"}
    }

    // Allowed methods
    if s := strings.TrimSpace(os.Getenv("CORS_ALLOWED_METHODS")); s != "" {
        parts := strings.Split(s, ",")
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }
        cfg.CORSAllowedMethods = parts
    } else {
        cfg.CORSAllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    }

    // Allowed headers
    if s := strings.TrimSpace(os.Getenv("CORS_ALLOWED_HEADERS")); s != "" {
        parts := strings.Split(s, ",")
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }
        cfg.CORSAllowedHeaders = parts
    } else {
        cfg.CORSAllowedHeaders = []string{"Content-Type", "Authorization", "X-API-Key", "X-API-Secret"}
    }

    // Exposed headers
    if s := strings.TrimSpace(os.Getenv("CORS_EXPOSE_HEADERS")); s != "" {
        parts := strings.Split(s, ",")
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }
        cfg.CORSExposedHeaders = parts
    } else {
        cfg.CORSExposedHeaders = []string{}
    }

    // Allow credentials
    switch strings.ToLower(strings.TrimSpace(os.Getenv("CORS_ALLOW_CREDENTIALS"))) {
    case "1", "true", "yes", "on":
        cfg.CORSAllowCredentials = true
    default:
        cfg.CORSAllowCredentials = false
    }

    // MaxAge
    if s := strings.TrimSpace(os.Getenv("CORS_MAX_AGE")); s != "" {
        if v, err := strconv.Atoi(s); err == nil && v >= 0 {
            cfg.CORSMaxAge = v
        } else {
            cfg.CORSMaxAge = 300
        }
    } else {
        cfg.CORSMaxAge = 300
    }

    // Validasi minimal
    if cfg.InfluxURL == "" || cfg.InfluxToken == "" || cfg.InfluxOrg == "" || cfg.InfluxBucket == "" {
        return Config{}, fmt.Errorf("missing required env: INFLUX_URL/INFLUX_TOKEN/INFLUX_ORG/INFLUX_BUCKET")
    }
    return cfg, nil
}