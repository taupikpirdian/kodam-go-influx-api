package usecase

// Layer: usecase (application)
// Peran: Mengorkestrasi logika aplikasi dan aturan bisnis tingkat use case.
// Bergantung pada domain, dan diekspose sebagai kontrak ke layer luar.

import (
    "context"

    "rti/influxdb/domain"
)

// HealthUsecase mendefinisikan kontrak untuk memeriksa status aplikasi.
type HealthUsecase interface {
    Check(ctx context.Context) (domain.HealthStatus, error)
}

// healthUsecase adalah implementasi sederhana dari HealthUsecase.
type healthUsecase struct{}

// NewHealthUsecase membuat instance HealthUsecase.
func NewHealthUsecase() HealthUsecase {
    return &healthUsecase{}
}

// Check melakukan pemeriksaan sederhana terhadap status aplikasi.
// Di dunia nyata, Anda bisa menambahkan pengecekan DB/redis/service lain.
func (u *healthUsecase) Check(ctx context.Context) (domain.HealthStatus, error) {
    return domain.HealthStatus{Status: "ok"}, nil
}