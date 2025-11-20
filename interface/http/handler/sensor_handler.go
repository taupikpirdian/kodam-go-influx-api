package handler

// Layer: interface (delivery)
// Peran: Mengelola HTTP endpoint untuk menyimpan data sensor personel.

import (
	"encoding/json"
	"fmt"
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
//
//	curl --location 'http://localhost:3000/api/sensors/personel' \
//	 --header 'Content-Type: application/json' \
//	 --data '{
//	     "timestamp": "2025-11-08T16:31:00Z",
//	     "client_code": "kodam",
//	     "json_data": "{\"person_id\":\"P12345\",\"name\":\"Sersan Budi2\",\"location\":{\"latitude\":-6.917464,\"longitude\":107.619123},\"status\":\"active\",\"heart_rate\":78,\"temperature\":36.5}"
//	 }'
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
			"timestamp":   ts.Format(time.RFC3339),
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

// PersonelData merepresentasikan struktur JSON terbaru yang tersimpan pada kolom json_data untuk sensor personel.
// Contoh payload:
//
//	{
//	  "generated_at": "2025-10-31T12:31:33.994Z",
//	  "reason": "gps",
//	  "personnel": {
//	    "6": { ... },
//	    "7": { ... }
//	  },
//	  "gps": {
//	    "gps-device-001": { ... }
//	  }
//	}
type PersonelData struct {
	GeneratedAt string `json:"generated_at"`
	Reason      string `json:"reason"`
	Personnel   map[string]struct {
		Timestamp    int64  `json:"timestamp"`
		SerialNumber string `json:"serial_number"`
		Identity     struct {
			ID        string `json:"id"`
			NRP       string `json:"nrp"`
			Name      string `json:"name"`
			Rank      string `json:"rank"`
			Unit      string `json:"unit"`
			Battalion string `json:"battalion"`
			Squad     string `json:"squad"`
			Avatar    string `json:"avatar"`
		} `json:"identity"`
		GPS struct {
			Latitude     float64 `json:"latitude"`
			Longitude    float64 `json:"longitude"`
			GPSTimestamp int64   `json:"gps_timestamp"`
		} `json:"gps"`
		RadioHealth struct {
			Heartrate          int   `json:"heartrate"`
			HeartrateTimestamp int64 `json:"heartrate_timestamp"`
		} `json:"radio_health"`
		Battery struct {
			Level int `json:"level"`
		} `json:"battery"`
		ReceivedAt int64 `json:"_received_at"`
	} `json:"personnel"`
	GPS map[string]struct {
		MessageType      string  `json:"message_type"`
		ReceivedAt       string  `json:"received_at"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
		FixQuality       int     `json:"fix_quality"`
		NumSatellites    int     `json:"num_satellites"`
		Hdop             float64 `json:"hdop"`
		AltitudeM        float64 `json:"altitude_m"`
		GeoidSeparationM float64 `json:"geoid_separation_m"`
		HeadingDeg       float64 `json:"heading_deg"`
	} `json:"gps"`
}

// personelListItem adalah item response untuk endpoint personel (streaming dan non-streaming)
// yang mengembalikan json_data sebagai objek terstruktur, bukan string mentah.
type personelListItem struct {
	Timestamp  string       `json:"timestamp"`
	ClientCode string       `json:"client_code"`
	JSONData   PersonelData `json:"json_data"`
}

// ADSBData merepresentasikan struktur JSON terbaru yang tersimpan pada kolom json_data untuk sensor ADSB.
// Contoh payload:
//
//	{
//	  "ts": "2025-11-19T06:11:49.593Z",
//	  "count": 77,
//	  "center": { "lat": -6.2088, "lon": 106.8456, "radius_m": 400000 },
//	  "flights": [
//	    {
//	      "id": "3d2a466b",
//	      "callsign": "GIA186",
//	      "icao24bit": null,
//	      "registration": "PK-GFH",
//	      "position": { "lat": -3.51, "lon": 103.986 },
//	      "altitude_ft": 34000,
//	      "heading_deg": 319,
//	      "groundSpeed_kts": 472,
//	      "verticalSpeed_fpm": 64,
//	      "updatedAt": "2025-11-19T06:11:49.591Z",
//	      "country": "Indonesia",
//	      "aircraftTypeICAO": "B738",
//	      "aircraftModelText": null,
//	      "aircraftTypeLabelDB": "Fixed-wing jet airliner",
//	      "images": null,
//	      "originAirport": null,
//	      "destinationAirport": null,
//	      "isOnGround": false,
//	      "fr24_link": "https://www.flightradar24.com/GIA186",
//	      "distance_km": 436.369,
//	      "source": "fr24"
//	    }
//	  ]
//	}
type ADSBData struct {
	Timestamp string `json:"ts"`
	Count     int    `json:"count"`
	Center    struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
		RadiusM   int     `json:"radius_m"`
	} `json:"center"`
	Flights []ADSBFlight `json:"flights"`
}

type ADSBFlight struct {
	ID           string  `json:"id"`
	Callsign     string  `json:"callsign"`
	ICAO24Bit    *string `json:"icao24bit"`
	Registration string  `json:"registration"`
	Position     struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
	} `json:"position"`
	AltitudeFt          int          `json:"altitude_ft"`
	HeadingDeg          int          `json:"heading_deg"`
	GroundSpeedKts      int          `json:"groundSpeed_kts"`
	VerticalSpeedFpm    int          `json:"verticalSpeed_fpm"`
	UpdatedAt           string       `json:"updatedAt"`
	Country             string       `json:"country"`
	AircraftTypeICAO    string       `json:"aircraftTypeICAO"`
	AircraftModelText   *string      `json:"aircraftModelText"`
	AircraftTypeLabelDB *string      `json:"aircraftTypeLabelDB"`
	Images              []ADSBImage  `json:"images"`
	OriginAirport       *ADSBAirport `json:"originAirport"`
	DestinationAirport  *ADSBAirport `json:"destinationAirport"`
	IsOnGround          bool         `json:"isOnGround"`
	FR24Link            string       `json:"fr24_link"`
	DistanceKm          float64      `json:"distance_km"`
	Source              string       `json:"source"`
}

type ADSBImage struct {
	Src       string `json:"src"`
	Link      string `json:"link"`
	Copyright string `json:"copyright"`
	Source    string `json:"source"`
}

type ADSBAirport struct {
	IATA      string  `json:"iata"`
	ICAO      string  `json:"icao"`
	Name      string  `json:"name"`
	ShortName *string `json:"shortName"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Position  struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
	} `json:"position"`
}

// adsbListItem adalah item response untuk endpoint ADSB (streaming dan non-streaming)
// yang mengembalikan json_data sebagai objek terstruktur, bukan string mentah.
type adsbListItem struct {
	Timestamp  string   `json:"timestamp"`
	ClientCode string   `json:"client_code"`
	JSONData   ADSBData `json:"json_data"`
}

// adsbListResponse adalah format response untuk endpoint ADSB non-streaming.
type adsbListResponse struct {
	Status  string           `json:"status"`
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    []adsbListItem   `json:"data"`
	Meta    listResponseMeta `json:"meta"`
}

// personelListResponse adalah format response untuk endpoint personel non-streaming.
type personelListResponse struct {
	Status  string             `json:"status"`
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []personelListItem `json:"data"`
	Meta    listResponseMeta   `json:"meta"`
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

	// Stream baris demi baris, mengubah json_data (string) menjadi objek PersonelData
	err := h.uc.StreamPersonelSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
		var pd PersonelData
		if err := json.Unmarshal([]byte(it.JSONData), &pd); err != nil {
			// Jangan matikan stream jika JSON tidak sesuai; log saja dan lanjut
			c.Logger().Warnf("[stream-personel] invalid json_data at %s (client_code=%s): %v; raw=%s",
				it.Timestamp.UTC().Format(time.RFC3339), it.ClientCode, err, it.JSONData)
			return nil
		}
		row := personelListItem{
			Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
			ClientCode: it.ClientCode,
			JSONData:   pd,
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

// GetPersonelList mengembalikan daftar data dalam format JSON array (non-streaming).
// Query params:
// - client_code (opsional)
// - start, stop (RFC3339, opsional)
// - page (default 1), limit (default 50)
// Contoh:
// curl --location 'http://localhost:3000/api/sensors/personel/list?client_code=kodam&start=2025-01-01T00:00:00Z&stop=2025-12-31T23:59:59Z&page=1&limit=50'
func (h *SensorHandler) GetPersonelList(c echo.Context) error {
	clientCode := strings.TrimSpace(c.QueryParam("client_code"))

	// Parse waktu opsional
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

	// Pagination params
	page := 1
	limit := 50
	if v := strings.TrimSpace(c.QueryParam("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	offset := (page - 1) * limit

	// Kumpulkan hasil menggunakan streaming tapi ditampung sebagai list dengan offset/limit
	var items []personelListItem
	idx := 0
	err := h.uc.StreamPersonelSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
		// Skip sampai offset
		if idx < offset {
			idx++
			return nil
		}
		if len(items) >= limit {
			return nil
		}
		var pd PersonelData
		if err := json.Unmarshal([]byte(it.JSONData), &pd); err != nil {
			return err
		}
		items = append(items, personelListItem{
			Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
			ClientCode: it.ClientCode,
			JSONData:   pd,
		})
		idx++
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, personelListResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    []personelListItem{},
			Meta:    listResponseMeta{Count: 0, Page: page, Limit: limit},
		})
	}

	return c.JSON(http.StatusOK, personelListResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "OK",
		Data:    items,
		Meta:    listResponseMeta{Count: len(items), Page: page, Limit: limit},
	})
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

// GetRadarList: versi non-streaming untuk radar
// Query params: client_code, start, stop (RFC3339), page (default 1), limit (default 50)
func (h *SensorHandler) GetRadarList(c echo.Context) error {
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

	page := 1
	limit := 50
	if v := strings.TrimSpace(c.QueryParam("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	offset := (page - 1) * limit

	var items []listItem
	idx := 0
	err := h.uc.StreamRadarSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
		if idx < offset {
			idx++
			return nil
		}
		if len(items) >= limit {
			return nil
		}
		items = append(items, listItem{
			Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
			ClientCode: it.ClientCode,
			JSONData:   it.JSONData,
		})
		idx++
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, listResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    []listItem{},
			Meta:    listResponseMeta{Count: 0, Page: page, Limit: limit},
		})
	}

	return c.JSON(http.StatusOK, listResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "OK",
		Data:    items,
		Meta:    listResponseMeta{Count: len(items), Page: page, Limit: limit},
	})
}

// PostDF menyimpan data ke measurement df_sensor.
// Payload sama dengan personel.
func (h *SensorHandler) PostDF(c echo.Context) error {
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

	if err := h.uc.StoreDFSensor(c.Request().Context(), input); err != nil {
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

// GetDF mengembalikan daftar data sensor df (streaming NDJSON).
// Mendukung filter waktu start/stop dan client_code.
func (h *SensorHandler) GetDF(c echo.Context) error {
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

	err := h.uc.StreamDFSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
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

// GetDFList: versi non-streaming untuk DF
// Query params: client_code, start, stop (RFC3339), page (default 1), limit (default 50)
func (h *SensorHandler) GetDFList(c echo.Context) error {
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

	page := 1
	limit := 50
	if v := strings.TrimSpace(c.QueryParam("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	offset := (page - 1) * limit

	var items []listItem
	idx := 0
	err := h.uc.StreamDFSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
		if idx < offset {
			idx++
			return nil
		}
		if len(items) >= limit {
			return nil
		}
		items = append(items, listItem{
			Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
			ClientCode: it.ClientCode,
			JSONData:   it.JSONData,
		})
		idx++
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, listResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    []listItem{},
			Meta:    listResponseMeta{Count: 0, Page: page, Limit: limit},
		})
	}

	return c.JSON(http.StatusOK, listResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "OK",
		Data:    items,
		Meta:    listResponseMeta{Count: len(items), Page: page, Limit: limit},
	})
}

// PostADSB menyimpan data ke measurement adsb_sensor.
// Payload sama dengan personel.
func (h *SensorHandler) PostADSB(c echo.Context) error {
	var req personelRequest
	// Decode JSON secara eksplisit agar tidak bergantung pada Content-Type header
	// yang kadang salah format saat pengujian dengan curl.
	dec := json.NewDecoder(c.Request().Body)
	if err := dec.Decode(&req); err != nil {
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

	if err := h.uc.StoreADSBSensor(c.Request().Context(), input); err != nil {
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

// GetADSB mengembalikan daftar data sensor adsb (streaming NDJSON).
// Mendukung filter waktu start/stop dan client_code.
func (h *SensorHandler) GetADSB(c echo.Context) error {
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

	// Set header untuk NDJSON streaming
	c.Response().Header().Set(echo.HeaderContentType, "application/x-ndjson")
	c.Response().WriteHeader(http.StatusOK)
	enc := json.NewEncoder(c.Response().Writer)

	// Stream baris demi baris, mengubah json_data (string) menjadi objek ADSBData
	err := h.uc.StreamADSBSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
		var ad ADSBData
		// Handle escaped JSON string
		jsonStr := it.JSONData
		if strings.HasPrefix(jsonStr, "\"") && strings.HasSuffix(jsonStr, "\"") {
			// Unescape the JSON string if it's wrapped in quotes
			var unescaped string
			if err := json.Unmarshal([]byte(jsonStr), &unescaped); err == nil {
				jsonStr = unescaped
			}
		}
		if err := json.Unmarshal([]byte(jsonStr), &ad); err != nil {
			fmt.Println("==========================")
			fmt.Println(err)
			// Jangan matikan stream jika JSON tidak sesuai; log saja dan lanjut
			c.Logger().Warnf("[stream-adsb] invalid json_data at %s (client_code=%s): %v; raw=%s",
				it.Timestamp.UTC().Format(time.RFC3339), it.ClientCode, err, it.JSONData)
			return nil
		}
		row := adsbListItem{
			Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
			ClientCode: it.ClientCode,
			JSONData:   ad,
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

// GetADSBList: versi non-streaming untuk ADSB
// Query params: client_code, start, stop (RFC3339), page (default 1), limit (default 50)
func (h *SensorHandler) GetADSBList(c echo.Context) error {
	clientCode := strings.TrimSpace(c.QueryParam("client_code"))

	// Parse waktu opsional
	var startPtr, stopPtr *time.Time
	if startStr := c.QueryParam("start"); startStr != "" {
		st, err := time.Parse(time.RFC3339, strings.TrimSpace(startStr))
		if err != nil {
			return c.JSON(http.StatusBadRequest, adsbListResponse{
				Status:  "error",
				Code:    http.StatusBadRequest,
				Message: "invalid start time format, use RFC3339",
				Data:    []adsbListItem{},
				Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
			})
		}
		startPtr = &st
	}
	if stopStr := c.QueryParam("stop"); stopStr != "" {
		sp, err := time.Parse(time.RFC3339, strings.TrimSpace(stopStr))
		if err != nil {
			return c.JSON(http.StatusBadRequest, adsbListResponse{
				Status:  "error",
				Code:    http.StatusBadRequest,
				Message: "invalid stop time format, use RFC3339",
				Data:    []adsbListItem{},
				Meta:    listResponseMeta{Count: 0, Page: 1, Limit: 0},
			})
		}
		stopPtr = &sp
	}

	// Pagination params
	page := 1
	limit := 50
	if v := strings.TrimSpace(c.QueryParam("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	offset := (page - 1) * limit

	// Kumpulkan hasil menggunakan streaming tapi ditampung sebagai list dengan offset/limit
	var items []adsbListItem
	idx := 0
	err := h.uc.StreamADSBSensors(c.Request().Context(), startPtr, stopPtr, clientCode, func(it domain.SensorInput) error {
		// Skip sampai offset
		if idx < offset {
			idx++
			return nil
		}
		if len(items) >= limit {
			return nil
		}
		var ad ADSBData
		// Handle escaped JSON string
		jsonStr := it.JSONData
		if strings.HasPrefix(jsonStr, "\"") && strings.HasSuffix(jsonStr, "\"") {
			// Unescape the JSON string if it's wrapped in quotes
			var unescaped string
			if err := json.Unmarshal([]byte(jsonStr), &unescaped); err == nil {
				jsonStr = unescaped
			}
		}
		if err := json.Unmarshal([]byte(jsonStr), &ad); err != nil {
			return err
		}
		items = append(items, adsbListItem{
			Timestamp:  it.Timestamp.UTC().Format(time.RFC3339),
			ClientCode: it.ClientCode,
			JSONData:   ad,
		})
		idx++
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, adsbListResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    []adsbListItem{},
			Meta:    listResponseMeta{Count: 0, Page: page, Limit: limit},
		})
	}

	return c.JSON(http.StatusOK, adsbListResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "OK",
		Data:    items,
		Meta:    listResponseMeta{Count: len(items), Page: page, Limit: limit},
	})
}

// availableDatesResponse adalah format response untuk endpoint available dates
type availableDatesResponse struct {
	Status  bool     `json:"status"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

// GetAvailableDatesPersonel mengembalikan daftar tanggal yang tersedia untuk data personel.
// Query params:
// - date (YYYY-MM-DD, required): tanggal untuk mencari data dalam bulan tersebut
// - client_code (required): kode client
// Contoh curl:
// curl --location 'http://localhost:3000/api/sensors/available-data/personel?date=2025-11-01&client_code=kodam'
func (h *SensorHandler) GetAvailableDatesPersonel(c echo.Context) error {
	dateStr := strings.TrimSpace(c.QueryParam("date"))
	clientCode := strings.TrimSpace(c.QueryParam("client_code"))

	if dateStr == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "date parameter is required (format: YYYY-MM-DD)",
			Data:    []string{},
		})
	}

	if clientCode == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "client_code parameter is required",
			Data:    []string{},
		})
	}

	// Parse date untuk mendapatkan awal dan akhir bulan
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "invalid date format, use YYYY-MM-DD",
			Data:    []string{},
		})
	}

	// Set start ke awal bulan dan stop ke akhir bulan
	year := date.Year()
	month := date.Month()
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month+1, 1, 0, 0, 0, -1, time.UTC) // akhir bulan saat ini

	dates, err := h.uc.GetAvailableDatesPersonel(c.Request().Context(), &start, &end, clientCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, availableDatesResponse{
			Status:  false,
			Message: err.Error(),
			Data:    []string{},
		})
	}

	return c.JSON(http.StatusOK, availableDatesResponse{
		Status:  true,
		Message: "success",
		Data:    dates,
	})
}

// GetAvailableDatesRadar mengembalikan daftar tanggal yang tersedia untuk data radar.
// Query params:
// - date (YYYY-MM-DD, required): tanggal untuk mencari data dalam bulan tersebut
// - client_code (required): kode client
// Contoh curl:
// curl --location 'http://localhost:3000/api/sensors/available-data/radar?date=2025-11-01&client_code=kodam'
func (h *SensorHandler) GetAvailableDatesRadar(c echo.Context) error {
	dateStr := strings.TrimSpace(c.QueryParam("date"))
	clientCode := strings.TrimSpace(c.QueryParam("client_code"))

	if dateStr == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "date parameter is required (format: YYYY-MM-DD)",
			Data:    []string{},
		})
	}

	if clientCode == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "client_code parameter is required",
			Data:    []string{},
		})
	}

	// Parse date untuk mendapatkan awal dan akhir bulan
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "invalid date format, use YYYY-MM-DD",
			Data:    []string{},
		})
	}

	// Set start ke awal bulan dan stop ke akhir bulan
	year := date.Year()
	month := date.Month()
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month+1, 1, 0, 0, 0, -1, time.UTC) // akhir bulan saat ini

	dates, err := h.uc.GetAvailableDatesRadar(c.Request().Context(), &start, &end, clientCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, availableDatesResponse{
			Status:  false,
			Message: err.Error(),
			Data:    []string{},
		})
	}

	return c.JSON(http.StatusOK, availableDatesResponse{
		Status:  true,
		Message: "success",
		Data:    dates,
	})
}

// GetAvailableDatesADSB mengembalikan daftar tanggal yang tersedia untuk data ADSB.
// Query params:
// - date (YYYY-MM-DD, required): tanggal untuk mencari data dalam bulan tersebut
// - client_code (required): kode client
// Contoh curl:
// curl --location 'http://localhost:3000/api/sensors/available-data/adsb?date=2025-11-01&client_code=kodam'
func (h *SensorHandler) GetAvailableDatesADSB(c echo.Context) error {
	dateStr := strings.TrimSpace(c.QueryParam("date"))
	clientCode := strings.TrimSpace(c.QueryParam("client_code"))

	if dateStr == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "date parameter is required (format: YYYY-MM-DD)",
			Data:    []string{},
		})
	}

	if clientCode == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "client_code parameter is required",
			Data:    []string{},
		})
	}

	// Parse date untuk mendapatkan awal dan akhir bulan
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "invalid date format, use YYYY-MM-DD",
			Data:    []string{},
		})
	}

	// Set start ke awal bulan dan stop ke akhir bulan
	year := date.Year()
	month := date.Month()
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month+1, 1, 0, 0, 0, -1, time.UTC) // akhir bulan saat ini

	dates, err := h.uc.GetAvailableDatesADSB(c.Request().Context(), &start, &end, clientCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, availableDatesResponse{
			Status:  false,
			Message: err.Error(),
			Data:    []string{},
		})
	}

	return c.JSON(http.StatusOK, availableDatesResponse{
		Status:  true,
		Message: "success",
		Data:    dates,
	})
}

// GetAvailableDatesDF mengembalikan daftar tanggal yang tersedia untuk data DF.
// Query params:
// - date (YYYY-MM-DD, required): tanggal untuk mencari data dalam bulan tersebut
// - client_code (required): kode client
// Contoh curl:
// curl --location 'http://localhost:3000/api/sensors/available-data/df?date=2025-11-01&client_code=kodam'
func (h *SensorHandler) GetAvailableDatesDF(c echo.Context) error {
	dateStr := strings.TrimSpace(c.QueryParam("date"))
	clientCode := strings.TrimSpace(c.QueryParam("client_code"))

	if dateStr == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "date parameter is required (format: YYYY-MM-DD)",
			Data:    []string{},
		})
	}

	if clientCode == "" {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "client_code parameter is required",
			Data:    []string{},
		})
	}

	// Parse date untuk mendapatkan awal dan akhir bulan
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, availableDatesResponse{
			Status:  false,
			Message: "invalid date format, use YYYY-MM-DD",
			Data:    []string{},
		})
	}

	// Set start ke awal bulan dan stop ke akhir bulan
	year := date.Year()
	month := date.Month()
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month+1, 1, 0, 0, 0, -1, time.UTC) // akhir bulan saat ini

	dates, err := h.uc.GetAvailableDatesDF(c.Request().Context(), &start, &end, clientCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, availableDatesResponse{
			Status:  false,
			Message: err.Error(),
			Data:    []string{},
		})
	}

	return c.JSON(http.StatusOK, availableDatesResponse{
		Status:  true,
		Message: "success",
		Data:    dates,
	})
}
