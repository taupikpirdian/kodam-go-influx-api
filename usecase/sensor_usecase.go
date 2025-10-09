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