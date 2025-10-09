package middleware

import (
    "net"
    "net/http"
    "sync"
    "time"

    "github.com/labstack/echo/v4"
    "golang.org/x/time/rate"

    "rti/influxdb/infrastructure/config"
)

// apiKeySecretMiddleware memeriksa header X-API-Key dan X-API-Secret jika dikonfigurasi.
func apiKeySecretMiddleware(cfg config.Config) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Jika tidak dikonfigurasi, lewati
            if cfg.APIKey == "" || cfg.APISecret == "" {
                return next(c)
            }
            key := c.Request().Header.Get("X-API-Key")
            secret := c.Request().Header.Get("X-API-Secret")
            if key != cfg.APIKey || secret != cfg.APISecret {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "success": false,
                    "status": http.StatusUnauthorized,
                    "status_message": http.StatusText(http.StatusUnauthorized),
                    "message": "invalid api credentials",
                    "data": nil,
                })
            }
            return next(c)
        }
    }
}

// ipWhitelistMiddleware membatasi akses berdasarkan daftar IP yang diizinkan.
func ipWhitelistMiddleware(cfg config.Config) echo.MiddlewareFunc {
    allowed := map[string]struct{}{}
    for _, ip := range cfg.AllowedIPs {
        if ip != "" {
            allowed[ip] = struct{}{}
        }
    }
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Jika whitelist tidak diaktifkan, lewati
            if !cfg.IPWhitelistEnabled {
                return next(c)
            }
            // Jika diaktifkan tetapi daftar kosong, tolak semua
            if len(allowed) == 0 {
                return c.JSON(http.StatusForbidden, map[string]interface{}{
                    "success": false,
                    "status": http.StatusForbidden,
                    "status_message": http.StatusText(http.StatusForbidden),
                    "message": "ip whitelist enabled but ALLOWED_IPS is empty",
                    "data": nil,
                })
            }
            rip := c.RealIP()
            // Normalize IPv6 loopback
            if rip == "::1" {
                rip = "127.0.0.1"
            }
            // Ekstrak host dari RemoteAddr jika diperlukan
            host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
            if err == nil && host != "" {
                rip = host
            }
            if _, ok := allowed[rip]; !ok {
                return c.JSON(http.StatusForbidden, map[string]interface{}{
                    "success": false,
                    "status": http.StatusForbidden,
                    "status_message": http.StatusText(http.StatusForbidden),
                    "message": "ip not allowed",
                    "data": nil,
                })
            }
            return next(c)
        }
    }
}

// rateLimitMiddleware membatasi jumlah request per menit berdasarkan API key atau IP.
func rateLimitMiddleware(cfg config.Config) echo.MiddlewareFunc {
    // Default 30 req/menit
    limit := cfg.RateLimitPerMinute
    if limit <= 0 {
        limit = 30
    }
    // Map limiter per identifier (apiKey atau IP)
    var (
        mu       sync.Mutex
        limiters = map[string]*rate.Limiter{}
    )
    // Helper membuat limiter per menit
    newLimiter := func() *rate.Limiter {
        // rate.NewLimiter dengan rate.Every(period) dan burst = limit
        return rate.NewLimiter(rate.Every(time.Minute/time.Duration(limit)), limit)
    }
    getLimiter := func(id string) *rate.Limiter {
        mu.Lock()
        defer mu.Unlock()
        l, ok := limiters[id]
        if !ok {
            l = newLimiter()
            limiters[id] = l
        }
        return l
    }
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Identitas: gunakan API key jika ada, jika tidak gunakan IP
            id := c.Request().Header.Get("X-API-Key")
            if id == "" {
                id = c.RealIP()
            }
            if id == "" {
                id = "anonymous"
            }
            limiter := getLimiter(id)
            if !limiter.Allow() {
                return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
                    "success": false,
                    "status": http.StatusTooManyRequests,
                    "status_message": http.StatusText(http.StatusTooManyRequests),
                    "message": "rate limit exceeded",
                    "data": nil,
                })
            }
            return next(c)
        }
    }
}

// SecurityMiddleware menggabungkan semua middleware keamanan.
func SecurityMiddleware(cfg config.Config) []echo.MiddlewareFunc {
    return []echo.MiddlewareFunc{
        apiKeySecretMiddleware(cfg),
        ipWhitelistMiddleware(cfg),
        rateLimitMiddleware(cfg),
    }
}