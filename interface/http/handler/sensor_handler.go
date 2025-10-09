package handler

// Layer: interface (delivery)
// Peran: Mengelola HTTP endpoint untuk menyimpan data sensor personel.

import (
    "encoding/json"
    "net/http"
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
    // Ambil filter client_code (opsional)
    clientCode := strings.TrimSpace(c.QueryParam("client_code"))

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
                Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
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
                Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
            })
        }
        stopPtr = &sp
    }

    // Set header untuk NDJSON streaming
    c.Response().Header().Set(echo.HeaderContentType, "application/x-ndjson")
    c.Response().WriteHeader(http.StatusOK)
    enc := json.NewEncoder(c.Response().Writer)

    // Stream baris demi baris
    err := h.uc.StreamPersonelSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
        row := listItem{
            Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
            ClientCode: it.ClientCode,
            JSONData:   it.JSONData,
        }
        if err := enc.Encode(row); err != nil {
            return err
        }
        if f, ok := c.Response().Writer.(http.Flusher); ok {
            f.Flush()
        }
        return nil
    })
    if err != nil {
        // Jika terjadi error sebelum ada data yang terkirim, kita kembalikan JSON error biasa.
        // Jika error terjadi setelah sebagian data terkirim, koneksi akan ditutup.
        if !c.Response().Committed {
            return c.JSON(http.StatusInternalServerError, listResponse{
                Status:  "error",
                Code:    http.StatusInternalServerError,
                Message: err.Error(),
                Data:    []listItem{},
                Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
            })
        }
    }
    return nil
}

// PostRadar menyimpan data ke measurement radar_sensor.
// Payload sama dengan personel.
func (h *SensorHandler) PostRadar(c echo.Context) error {
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

    if err := h.uc.StoreRadarSensor(c.Request().Context(), input); err != nil {
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

// GetRadar mengembalikan daftar data sensor radar (streaming NDJSON).
// Mendukung filter waktu start/stop dan client_code.
func (h *SensorHandler) GetRadar(c echo.Context) error {
    clientCode := strings.TrimSpace(c.QueryParam("client_code"))

    var startPtr, stopPtr *time.Time
    if startStr := c.QueryParam("start"); startStr != "" {
        st, err := time.Parse(time.RFC3339, strings.TrimSpace(startStr))
        if err != nil {
            return c.JSON(http.StatusBadRequest, listResponse{
                Status:  "error",
                Code:    http.StatusBadRequest,
                Message: "invalid start time format, use RFC3339",
                Data:    []listItem{},
                Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
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
                Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
            })
        }
        stopPtr = &sp
    }

    c.Response().Header().Set(echo.HeaderContentType, "application/x-ndjson")
    c.Response().WriteHeader(http.StatusOK)
    enc := json.NewEncoder(c.Response().Writer)

    err := h.uc.StreamRadarSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
        row := listItem{
            Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
            ClientCode: it.ClientCode,
            JSONData:   it.JSONData,
        }
        if err := enc.Encode(row); err != nil {
            return err
        }
        if f, ok := c.Response().Writer.(http.Flusher); ok {
            f.Flush()
        }
        return nil
    })
    if err != nil {
        if !c.Response().Committed {
            return c.JSON(http.StatusInternalServerError, listResponse{
                Status:  "error",
                Code:    http.StatusInternalServerError,
                Message: err.Error(),
                Data:    []listItem{},
                Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
            })
        }
    }
    return nil
}