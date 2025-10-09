package handler

// Layer: interface (delivery)
// Peran: Mengelola HTTP endpoint untuk menyimpan data sensor personel.

import (
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/labstack/echo/v4"

    "rti/influxdb/domain"
    "rti/influxdb/usecase"
)

// SensorHandler menangani endpoint terkait sensor personel.
type SensorHandler struct {
    uc usecase.SensorUsecase
}

// NewSensorHandler membuat instance SensorHandler.
func NewSensorHandler(uc usecase.SensorUsecase) *SensorHandler {
    return &SensorHandler{uc: uc}
}

// apiResponse adalah format response standar sesuai permintaan.
type apiResponse struct {
    Success       bool        `json:"success"`
    Status        int         `json:"status"`
    StatusMessage string      `json:"status_message"`
    Message       string      `json:"message"`
    Data          interface{} `json:"data"`
}

// personelRequest merepresentasikan payload POST /api/sensors/personel.
type personelRequest struct {
    Timestamp  string `json:"timestamp"`
    ClientCode string `json:"client_code"`
    JSONData   string `json:"json_data"`
}

// PostPersonel menyimpan data ke measurement personel_sensor.
// Contoh curl:
// curl --location 'http://localhost:3000/api/sensors/personel' \
//  --header 'Content-Type: application/json' \
//  --data '{
//      "timestamp": "2025-11-08T16:31:00Z",
//      "client_code": "kodam",
//      "json_data": "{\"person_id\":\"P12345\",\"name\":\"Sersan Budi2\",\"location\":{\"latitude\":-6.917464,\"longitude\":107.619123},\"status\":\"active\",\"heart_rate\":78,\"temperature\":36.5}"
//  }'
func (h *SensorHandler) PostPersonel(c echo.Context) error {
    var req personelRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, apiResponse{
            Success:       false,
            Status:        http.StatusBadRequest,
            StatusMessage: http.StatusText(http.StatusBadRequest),
            Message:       "invalid request body",
            Data:          nil,
        })
    }

    ts, err := time.Parse(time.RFC3339, req.Timestamp)
    if err != nil {
        return c.JSON(http.StatusBadRequest, apiResponse{
            Success:       false,
            Status:        http.StatusBadRequest,
            StatusMessage: http.StatusText(http.StatusBadRequest),
            Message:       "invalid timestamp format, use RFC3339",
            Data:          nil,
        })
    }

    input := domain.SensorInput{
        Timestamp:  ts,
        ClientCode: req.ClientCode,
        JSONData:   req.JSONData,
    }

    if err := h.uc.StorePersonelSensor(c.Request().Context(), input); err != nil {
        return c.JSON(http.StatusInternalServerError, apiResponse{
            Success:       false,
            Status:        http.StatusInternalServerError,
            StatusMessage: http.StatusText(http.StatusInternalServerError),
            Message:       err.Error(),
            Data:          nil,
        })
    }

    return c.JSON(http.StatusOK, apiResponse{
        Success:       true,
        Status:        http.StatusOK,
        StatusMessage: http.StatusText(http.StatusOK),
        Message:       "Success",
        Data:          nil,
    })
}

// TestStore adalah endpoint sederhana untuk menulis data contoh ke InfluxDB.
// Endpoint: POST /api/test/influx
// Menggunakan nilai default (timestamp: now, client_code: "test", json_data: sample JSON).
// Cocok untuk sanity-check koneksi dan pipeline repository/usecase.
// Contoh curl:
// curl --location 'http://localhost:3000/api/test/influx' --request POST
func (h *SensorHandler) TestStore(c echo.Context) error {
    ts := time.Now().UTC()
    input := domain.SensorInput{
        Timestamp:  ts,
        ClientCode: "test",
        JSONData:   "{\"ping\":\"ok\",\"note\":\"simple test write\"}",
    }

    if err := h.uc.StorePersonelSensor(c.Request().Context(), input); err != nil {
        return c.JSON(http.StatusInternalServerError, apiResponse{
            Success:       false,
            Status:        http.StatusInternalServerError,
            StatusMessage: http.StatusText(http.StatusInternalServerError),
            Message:       err.Error(),
            Data:          nil,
        })
    }

    return c.JSON(http.StatusOK, apiResponse{
        Success:       true,
        Status:        http.StatusOK,
        StatusMessage: http.StatusText(http.StatusOK),
        Message:       "Test point written",
        Data: map[string]string{
            "timestamp":  ts.Format(time.RFC3339),
            "client_code": "test",
            "json_data":   "{\"ping\":\"ok\",\"note\":\"simple test write\"}",
        },
    })
}

// listResponse adalah format response untuk GET list sesuai permintaan.
type listResponse struct {
    Status  string           `json:"status"`
    Code    int              `json:"code"`
    Message string           `json:"message"`
    Data    []listItem       `json:"data"`
    Meta    listResponseMeta `json:"meta"`
}

type listItem struct {
    Timestamp  string `json:"timestamp"`
    ClientCode string `json:"client_code"`
    JSONData   string `json:"json_data"`
}

type listResponseMeta struct {
    Count int `json:"count"`
    Page  int `json:"page"`
    Limit int `json:"limit"`
}

// GetPersonel mengembalikan daftar data sensor yang tersimpan.
// Contoh curl:
// curl --location 'http://localhost:3000/api/sensors/personel'
func (h *SensorHandler) GetPersonel(c echo.Context) error {
    // Ambil query param page & limit (opsional)
    page := 1
    limit := 50
    if p := c.QueryParam("page"); p != "" {
        if v, err := strconv.Atoi(p); err == nil && v > 0 {
            page = v
        }
    }
    if l := c.QueryParam("limit"); l != "" {
        if v, err := strconv.Atoi(l); err == nil && v > 0 {
            limit = v
        }
    }

    // Ambil filter waktu (opsional) dengan format RFC3339: 2006-01-02T15:04:05Z
    var startPtr, stopPtr *time.Time
    if startStr := c.QueryParam("start"); startStr != "" {
        st, err := time.Parse(time.RFC3339, strings.TrimSpace(startStr))
        if err != nil {
            return c.JSON(http.StatusBadRequest, listResponse{
                Status:  "error",
                Code:    http.StatusBadRequest,
                Message: "invalid start time format, use RFC3339",
                Data:    []listItem{},
                Meta:    listResponseMeta{Count: 0, Page: page, Limit: limit},
            })
        }
        startPtr = &st
    }
    if stopStr := c.QueryParam("stop"); stopStr != "" {
        sp, err := time.Parse(time.RFC3339, strings.TrimSpace(stopStr))
        if err != nil {
            return c.JSON(http.StatusBadRequest, listResponse{
                Status:  "error",
                Code:    http.StatusBadRequest,
                Message: "invalid stop time format, use RFC3339",
                Data:    []listItem{},
                Meta:    listResponseMeta{Count: 0, Page: page, Limit: limit},
            })
        }
        stopPtr = &sp
    }

    items, err := h.uc.FetchPersonelSensors(c.Request().Context(), page, limit, startPtr, stopPtr)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, listResponse{
            Status:  "error",
            Code:    http.StatusInternalServerError,
            Message: err.Error(),
            Data:    []listItem{},
            Meta:    listResponseMeta{Count: 0, Page: page, Limit: limit},
        })
    }

    // Map domain ke response
    respItems := make([]listItem, 0, len(items))
    for _, it := range items {
        respItems = append(respItems, listItem{
            Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
            ClientCode: it.ClientCode,
            JSONData:   it.JSONData,
        })
    }

    return c.JSON(http.StatusOK, listResponse{
        Status:  "success",
        Code:    http.StatusOK,
        Message: "Sensor data list retrieved successfully",
        Data:    respItems,
        Meta:    listResponseMeta{Count: len(respItems), Page: page, Limit: limit},
    })
}