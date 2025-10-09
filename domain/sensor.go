package domain

// Layer: domain
// Peran: Mendefinisikan entitas/tipe inti yang merepresentasikan data
// personel sensor yang akan disimpan ke InfluxDB.

import "time"

// SensorInput merepresentasikan payload yang diterima dari API untuk disimpan.
type SensorInput struct {
    Timestamp  time.Time
    ClientCode string
    JSONData   string
}