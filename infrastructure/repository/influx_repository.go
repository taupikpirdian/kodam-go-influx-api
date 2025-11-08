package repository

// Layer: infrastructure (repository)
// Peran: Implementasi detail penyimpanan ke InfluxDB.

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"rti/influxdb/domain"
	"rti/influxdb/infrastructure/config"
)

// InfluxRepository menyimpan koneksi dan write API ke InfluxDB.
type InfluxRepository struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	org      string
	bucket   string
}

// NewInfluxRepository membuat repository untuk menulis data ke InfluxDB.
func NewInfluxRepository(cfg config.Config) *InfluxRepository {
	client := influxdb2.NewClient(cfg.InfluxURL, cfg.InfluxToken)
	writeAPI := client.WriteAPIBlocking(cfg.InfluxOrg, cfg.InfluxBucket)
	return &InfluxRepository{
		client:   client,
		writeAPI: writeAPI,
		org:      cfg.InfluxOrg,
		bucket:   cfg.InfluxBucket,
	}
}

// Close menutup client InfluxDB.
func (r *InfluxRepository) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

// CheckConnection melakukan ping ke server InfluxDB. Jika gagal, kembalikan error.
func (r *InfluxRepository) CheckConnection(ctx context.Context) error {
	ok, err := r.client.Ping(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("failed to ping InfluxDB")
	}
	return nil
}

// WritePersonelSensor menulis data ke measurement personel_sensor dengan tag client_code
// dan field json_data.
func (r *InfluxRepository) WritePersonelSensor(ctx context.Context, input domain.SensorInput) error {
	p := influxdb2.NewPoint(
		"personel_sensor",
		map[string]string{
			"client_code": input.ClientCode,
		},
		map[string]interface{}{
			"json_data": input.JSONData,
		},
		input.Timestamp,
	)
	return r.writeAPI.WritePoint(ctx, p)
}

// WriteRadarSensor menulis data ke measurement radar_sensor dengan tag client_code
// dan field json_data.
func (r *InfluxRepository) WriteRadarSensor(ctx context.Context, input domain.SensorInput) error {
	p := influxdb2.NewPoint(
		"radar_sensor",
		map[string]string{
			"client_code": input.ClientCode,
		},
		map[string]interface{}{
			"json_data": input.JSONData,
		},
		input.Timestamp,
	)
	return r.writeAPI.WritePoint(ctx, p)
}

// QueryPersonelSensor mengambil daftar data sensor dari measurement personel_sensor.
// Mendukung pagination sederhana (page, limit), filter range waktu (start, stop), dan mengurutkan berdasarkan waktu terbaru.
func (r *InfluxRepository) QueryPersonelSensor(ctx context.Context, page, limit int, start, stop *time.Time) ([]domain.SensorInput, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Bangun bagian range waktu secara dinamis.
	var rangeClause string
	switch {
	case start != nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: %s))", start.UTC().Format(time.RFC3339), stop.UTC().Format(time.RFC3339))
	case start != nil && stop == nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: 2100-01-01T00:00:00Z))", start.UTC().Format(time.RFC3339))
	case start == nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: 0), stop: time(v: %s))", stop.UTC().Format(time.RFC3339))
	default:
		// Default: rentang sangat luas termasuk masa depan
		rangeClause = "|> range(start: time(v: 0), stop: time(v: 2100-01-01T00:00:00Z))"
	}

	flux := fmt.Sprintf(`from(bucket: %q)
        %s
        |> filter(fn: (r) => r._measurement == "personel_sensor" and r._field == "json_data")
        |> keep(columns: ["_time", "client_code", "_value"])
        |> sort(columns: ["_time"], desc: true)
        |> limit(n: %d, offset: %d)`, r.bucket, rangeClause, limit, offset)

	q := r.client.QueryAPI(r.org)
	res, err := q.Query(ctx, flux)
	if err != nil {
		return nil, err
	}

	var list []domain.SensorInput
	for res.Next() {
		rec := res.Record()
		ts := rec.Time()
		// client_code adalah tag
		var clientCode string
		if v := rec.ValueByKey("client_code"); v != nil {
			if s, ok := v.(string); ok {
				clientCode = s
			}
		}
		// _value adalah field json_data
		var jsonData string
		if v := rec.Value(); v != nil {
			if s, ok := v.(string); ok {
				jsonData = s
			} else {
				jsonData = fmt.Sprintf("%v", v)
			}
		}
		list = append(list, domain.SensorInput{
			Timestamp:  ts,
			ClientCode: clientCode,
			JSONData:   jsonData,
		})
	}
	if res.Err() != nil {
		return nil, res.Err()
	}
	return list, nil
}

// StreamPersonelSensor melakukan query dan mem-stream setiap baris hasil melalui callback onRow.
// Mendukung filter opsional client_code.
// Cocok untuk data berukuran besar agar tidak menampung semua data di memori.
func (r *InfluxRepository) StreamPersonelSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
	// Bangun bagian range waktu secara dinamis.
	var rangeClause string
	switch {
	case start != nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: %s))", start.UTC().Format(time.RFC3339), stop.UTC().Format(time.RFC3339))
	case start != nil && stop == nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: 2100-01-01T00:00:00Z))", start.UTC().Format(time.RFC3339))
	case start == nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: 0), stop: time(v: %s))", stop.UTC().Format(time.RFC3339))
	default:
		// Default: rentang sangat luas termasuk masa depan
		rangeClause = "|> range(start: time(v: 0), stop: time(v: 2100-01-01T00:00:00Z))"
	}

	// Bangun filter bertahap agar lebih robust.
	var clientFilterStep string
	if clientCode != "" {
		clientFilterStep = fmt.Sprintf("\n        |> filter(fn: (r) => r.client_code == %q)", clientCode)
	}

	flux := fmt.Sprintf(`from(bucket: %q)
        %s
        |> filter(fn: (r) => r._measurement == "personel_sensor")
        |> filter(fn: (r) => r._field == "json_data")%s
        |> keep(columns: ["_time", "client_code", "_value"])
        |> sort(columns: ["_time"], desc: true)`, r.bucket, rangeClause, clientFilterStep)

	q := r.client.QueryAPI(r.org)
	res, err := q.Query(ctx, flux)
	if err != nil {
		return err
	}

	for res.Next() {
		rec := res.Record()
		ts := rec.Time()
		// client_code adalah tag
		var clientCode string
		if v := rec.ValueByKey("client_code"); v != nil {
			if s, ok := v.(string); ok {
				clientCode = s
			}
		}
		// _value adalah field json_data
		var jsonData string
		if v := rec.Value(); v != nil {
			if s, ok := v.(string); ok {
				jsonData = s
			} else {
				jsonData = fmt.Sprintf("%v", v)
			}
		}

		if err := onRow(domain.SensorInput{
			Timestamp:  ts,
			ClientCode: clientCode,
			JSONData:   jsonData,
		}); err != nil {
			return err
		}
	}
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

// WriteDFSensor menulis data ke measurement df_sensor dengan tag client_code
// dan field json_data.
func (r *InfluxRepository) WriteDFSensor(ctx context.Context, input domain.SensorInput) error {
	p := influxdb2.NewPoint(
		"df_sensor",
		map[string]string{
			"client_code": input.ClientCode,
		},
		map[string]interface{}{
			"json_data": input.JSONData,
		},
		input.Timestamp,
	)
	return r.writeAPI.WritePoint(ctx, p)
}

// StreamDFSensor melakukan query dan mem-stream setiap baris hasil melalui callback onRow.
// Mendukung filter opsional client_code.
func (r *InfluxRepository) StreamDFSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
	var rangeClause string
	switch {
	case start != nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: %s))", start.UTC().Format(time.RFC3339), stop.UTC().Format(time.RFC3339))
	case start != nil && stop == nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: 2100-01-01T00:00:00Z))", start.UTC().Format(time.RFC3339))
	case start == nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: 0), stop: time(v: %s))", stop.UTC().Format(time.RFC3339))
	default:
		rangeClause = "|> range(start: time(v: 0), stop: time(v: 2100-01-01T00:00:00Z))"
	}

	var clientFilterStep string
	if clientCode != "" {
		clientFilterStep = fmt.Sprintf("\n        |> filter(fn: (r) => r.client_code == %q)", clientCode)
	}

	flux := fmt.Sprintf(`from(bucket: %q)
        %s
        |> filter(fn: (r) => r._measurement == "df_sensor")
        |> filter(fn: (r) => r._field == "json_data")%s
        |> keep(columns: ["_time", "client_code", "_value"])
        |> sort(columns: ["_time"], desc: true)`, r.bucket, rangeClause, clientFilterStep)

	q := r.client.QueryAPI(r.org)
	res, err := q.Query(ctx, flux)
	if err != nil {
		return err
	}

	for res.Next() {
		rec := res.Record()
		ts := rec.Time()
		var cc string
		if v := rec.ValueByKey("client_code"); v != nil {
			if s, ok := v.(string); ok {
				cc = s
			}
		}
		var jsonData string
		if v := rec.Value(); v != nil {
			if s, ok := v.(string); ok {
				jsonData = s
			} else {
				jsonData = fmt.Sprintf("%v", v)
			}
		}

		if err := onRow(domain.SensorInput{Timestamp: ts, ClientCode: cc, JSONData: jsonData}); err != nil {
			return err
		}
	}
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

// WriteADSBSensor menulis data ke measurement adsb_sensor dengan tag client_code
// dan field json_data.
func (r *InfluxRepository) WriteADSBSensor(ctx context.Context, input domain.SensorInput) error {
	p := influxdb2.NewPoint(
		"adsb_sensor",
		map[string]string{
			"client_code": input.ClientCode,
		},
		map[string]interface{}{
			"json_data": input.JSONData,
		},
		input.Timestamp,
	)
	return r.writeAPI.WritePoint(ctx, p)
}

// StreamADSBSensor melakukan query dan mem-stream setiap baris hasil melalui callback onRow.
// Mendukung filter opsional client_code.
func (r *InfluxRepository) StreamADSBSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
	var rangeClause string
	switch {
	case start != nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: %s))", start.UTC().Format(time.RFC3339), stop.UTC().Format(time.RFC3339))
	case start != nil && stop == nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: 2100-01-01T00:00:00Z))", start.UTC().Format(time.RFC3339))
	case start == nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: 0), stop: time(v: %s))", stop.UTC().Format(time.RFC3339))
	default:
		rangeClause = "|> range(start: time(v: 0), stop: time(v: 2100-01-01T00:00:00Z))"
	}

	var clientFilterStep string
	if clientCode != "" {
		clientFilterStep = fmt.Sprintf("\n        |> filter(fn: (r) => r.client_code == %q)", clientCode)
	}

	flux := fmt.Sprintf(`from(bucket: %q)
        %s
        |> filter(fn: (r) => r._measurement == "adsb_sensor")
        |> filter(fn: (r) => r._field == "json_data")%s
        |> keep(columns: ["_time", "client_code", "_value"])
        |> sort(columns: ["_time"], desc: true)`, r.bucket, rangeClause, clientFilterStep)

	q := r.client.QueryAPI(r.org)
	res, err := q.Query(ctx, flux)
	if err != nil {
		return err
	}

	for res.Next() {
		rec := res.Record()
		ts := rec.Time()
		var cc string
		if v := rec.ValueByKey("client_code"); v != nil {
			if s, ok := v.(string); ok {
				cc = s
			}
		}
		var jsonData string
		if v := rec.Value(); v != nil {
			if s, ok := v.(string); ok {
				jsonData = s
			} else {
				jsonData = fmt.Sprintf("%v", v)
			}
		}

		if err := onRow(domain.SensorInput{Timestamp: ts, ClientCode: cc, JSONData: jsonData}); err != nil {
			return err
		}
	}
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

// StreamRadarSensor melakukan query dan mem-stream setiap baris hasil melalui callback onRow.
// Mendukung filter opsional client_code.
func (r *InfluxRepository) StreamRadarSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
	// Bangun bagian range waktu secara dinamis.
	var rangeClause string
	switch {
	case start != nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: %s))", start.UTC().Format(time.RFC3339), stop.UTC().Format(time.RFC3339))
	case start != nil && stop == nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: 2100-01-01T00:00:00Z))", start.UTC().Format(time.RFC3339))
	case start == nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: 0), stop: time(v: %s))", stop.UTC().Format(time.RFC3339))
	default:
		rangeClause = "|> range(start: time(v: 0), stop: time(v: 2100-01-01T00:00:00Z))"
	}

	var clientFilterStep string
	if clientCode != "" {
		clientFilterStep = fmt.Sprintf("\n        |> filter(fn: (r) => r.client_code == %q)", clientCode)
	}

	flux := fmt.Sprintf(`from(bucket: %q)
        %s
        |> filter(fn: (r) => r._measurement == "radar_sensor")
        |> filter(fn: (r) => r._field == "json_data")%s
        |> keep(columns: ["_time", "client_code", "_value"])
        |> sort(columns: ["_time"], desc: true)`, r.bucket, rangeClause, clientFilterStep)

	q := r.client.QueryAPI(r.org)
	res, err := q.Query(ctx, flux)
	if err != nil {
		return err
	}

	for res.Next() {
		rec := res.Record()
		ts := rec.Time()
		var cc string
		if v := rec.ValueByKey("client_code"); v != nil {
			if s, ok := v.(string); ok {
				cc = s
			}
		}
		var jsonData string
		if v := rec.Value(); v != nil {
			if s, ok := v.(string); ok {
				jsonData = s
			} else {
				jsonData = fmt.Sprintf("%v", v)
			}
		}

		if err := onRow(domain.SensorInput{Timestamp: ts, ClientCode: cc, JSONData: jsonData}); err != nil {
			return err
		}
	}
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

// GetAvailableDatesPersonel mengambil daftar tanggal yang tersedia untuk data personel sensor.
// Menggunakan Flux query untuk mendapatkan tanggal-tanggal unik dari measurement personel_sensor.
func (r *InfluxRepository) GetAvailableDatesPersonel(ctx context.Context, start, stop *time.Time, clientCode string) ([]string, error) {
	// Bangun range clause
	var rangeClause string
	switch {
	case start != nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: %s))", start.UTC().Format(time.RFC3339), stop.UTC().Format(time.RFC3339))
	case start != nil && stop == nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: %s), stop: time(v: 2100-01-01T00:00:00Z))", start.UTC().Format(time.RFC3339))
	case start == nil && stop != nil:
		rangeClause = fmt.Sprintf("|> range(start: time(v: 0), stop: time(v: %s))", stop.UTC().Format(time.RFC3339))
	default:
		rangeClause = "|> range(start: time(v: 0), stop: time(v: 2100-01-01T00:00:00Z))"
	}

	var clientFilterStep string
	if clientCode != "" {
		clientFilterStep = fmt.Sprintf("\n        |> filter(fn: (r) => r.client_code == %q)", clientCode)
	}

	// Flux query untuk mendapatkan tanggal unik dari personel_sensor
	flux := fmt.Sprintf(`from(bucket: %q)
        %s
        |> filter(fn: (r) => r._measurement == "personel_sensor")
        |> filter(fn: (r) => r._field == "json_data")%s
        |> map(fn: (r) => ({
            _time: r._time,
            _value: r._value,
            _date: date(t: r._time)
        }))
        |> group(columns: ["_date"])
        |> distinct(column: "_date")
        |> keep(columns: ["_date"])
        |> sort(columns: ["_date"], desc: false)`, r.bucket, rangeClause, clientFilterStep)

	q := r.client.QueryAPI(r.org)
	res, err := q.Query(ctx, flux)
	if err != nil {
		return nil, err
	}

	var dates []string
	for res.Next() {
		rec := res.Record()
		if dateValue := rec.Value(); dateValue != nil {
			// Convert date to string format YYYY-MM-DD
			if t, ok := dateValue.(time.Time); ok {
				dates = append(dates, t.Format("2006-01-02"))
			}
		}
	}

	if res.Err() != nil {
		return nil, res.Err()
	}

	return dates, nil
}
