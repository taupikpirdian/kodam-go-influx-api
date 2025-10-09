package domain

// Layer: domain
// Peran: Mewakili entitas/tipe inti bisnis. Bebas dari ketergantungan
// pada framework ataupun detail implementasi.

// HealthStatus merepresentasikan status kesehatan aplikasi.
type HealthStatus struct {
    Status string `json:"status"`
}