package handler

// Layer: interface (delivery)
// Peran: Menangani I/O (HTTP), memetakan request/response, dan memanggil usecase.
// Tidak berisi logika bisnis, hanya adaptor.

import (
    "net/http"

    "github.com/labstack/echo/v4"

    "rti/influxdb/usecase"
)

// HealthHandler adalah HTTP handler untuk endpoint health.
type HealthHandler struct {
    uc usecase.HealthUsecase
}

// NewHealthHandler membuat instance HealthHandler.
func NewHealthHandler(uc usecase.HealthUsecase) *HealthHandler {
    return &HealthHandler{uc: uc}
}

// Health adalah handler untuk GET /health.
// Contoh response JSON:
// {
//   "status": "ok"
// }
func (h *HealthHandler) Health(c echo.Context) error {
    status, err := h.uc.Check(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"status": "error"})
    }
    return c.JSON(http.StatusOK, status)
}