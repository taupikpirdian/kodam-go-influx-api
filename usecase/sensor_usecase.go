package usecase

// Layer: usecase (application)
// Peran: Koordinasi logika aplikasi untuk menyimpan data sensor personel.

import (
    "context"
    "errors"
    "time"

    "rti/influxdb/domain"
)

// SensorRepository mendefinisikan kontrak penyimpanan data sensor.
type SensorRepository interface {
    WritePersonelSensor(ctx context.Context, input domain.SensorInput) error
    QueryPersonelSensor(ctx context.Context, page, limit int, start, stop *time.Time) ([]domain.SensorInput, error)
    StreamPersonelSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error
    GetAvailableDatesPersonel(ctx context.Context, start, stop *time.Time, clientCode string) ([]string, error)
    // Radar
    WriteRadarSensor(ctx context.Context, input domain.SensorInput) error
    StreamRadarSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error
    // DF
    WriteDFSensor(ctx context.Context, input domain.SensorInput) error
    StreamDFSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error
    // ADSB
    WriteADSBSensor(ctx context.Context, input domain.SensorInput) error
    StreamADSBSensor(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error
}

// SensorUsecase mengorkestrasi penyimpanan data sensor.
type SensorUsecase struct {
    repo SensorRepository
}

// NewSensorUsecase membuat instance SensorUsecase.
func NewSensorUsecase(repo SensorRepository) SensorUsecase {
    return SensorUsecase{repo: repo}
}

// StorePersonelSensor melakukan validasi sederhana lalu menyimpan data ke repository.
func (u SensorUsecase) StorePersonelSensor(ctx context.Context, input domain.SensorInput) error {
    if input.ClientCode == "" {
        return errors.New("client_code is required")
    }
    if input.JSONData == "" {
        return errors.New("json_data is required")
    }
    if input.Timestamp.IsZero() {
        return errors.New("timestamp is required")
    }
    return u.repo.WritePersonelSensor(ctx, input)
}

// FetchPersonelSensors mengambil daftar data sensor untuk keperluan listing.
func (u SensorUsecase) FetchPersonelSensors(ctx context.Context, page, limit int, start, stop *time.Time) ([]domain.SensorInput, error) {
    return u.repo.QueryPersonelSensor(ctx, page, limit, start, stop)
}

// StreamPersonelSensors mem-forward streaming dari repository.
func (u SensorUsecase) StreamPersonelSensors(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
    return u.repo.StreamPersonelSensor(ctx, start, stop, clientCode, onRow)
}

// StoreRadarSensor melakukan validasi sederhana lalu menyimpan data ke repository radar.
func (u SensorUsecase) StoreRadarSensor(ctx context.Context, input domain.SensorInput) error {
    if input.ClientCode == "" {
        return errors.New("client_code is required")
    }
    if input.JSONData == "" {
        return errors.New("json_data is required")
    }
    if input.Timestamp.IsZero() {
        return errors.New("timestamp is required")
    }
    return u.repo.WriteRadarSensor(ctx, input)
}

// StreamRadarSensors mem-forward streaming dari repository radar.
func (u SensorUsecase) StreamRadarSensors(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
    return u.repo.StreamRadarSensor(ctx, start, stop, clientCode, onRow)
}

// StoreDFSensor melakukan validasi sederhana lalu menyimpan data DF ke repository.
func (u SensorUsecase) StoreDFSensor(ctx context.Context, input domain.SensorInput) error {
    if input.ClientCode == "" {
        return errors.New("client_code is required")
    }
    if input.JSONData == "" {
        return errors.New("json_data is required")
    }
    if input.Timestamp.IsZero() {
        return errors.New("timestamp is required")
    }
    return u.repo.WriteDFSensor(ctx, input)
}

// StreamDFSensors mem-forward streaming dari repository DF.
func (u SensorUsecase) StreamDFSensors(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
    return u.repo.StreamDFSensor(ctx, start, stop, clientCode, onRow)
}

// StoreADSBSensor melakukan validasi sederhana lalu menyimpan data ADSB ke repository.
func (u SensorUsecase) StoreADSBSensor(ctx context.Context, input domain.SensorInput) error {
    if input.ClientCode == "" {
        return errors.New("client_code is required")
    }
    if input.JSONData == "" {
        return errors.New("json_data is required")
    }
    if input.Timestamp.IsZero() {
        return errors.New("timestamp is required")
    }
    return u.repo.WriteADSBSensor(ctx, input)
}

// StreamADSBSensors mem-forward streaming dari repository ADSB.
func (u SensorUsecase) StreamADSBSensors(ctx context.Context, start, stop *time.Time, clientCode string, onRow func(domain.SensorInput) error) error {
    return u.repo.StreamADSBSensor(ctx, start, stop, clientCode, onRow)
}

// GetAvailableDatesPersonel mengambil daftar tanggal yang tersedia untuk data personel.
func (u SensorUsecase) GetAvailableDatesPersonel(ctx context.Context, start, stop *time.Time, clientCode string) ([]string, error) {
    return u.repo.GetAvailableDatesPersonel(ctx, start, stop, clientCode)
}