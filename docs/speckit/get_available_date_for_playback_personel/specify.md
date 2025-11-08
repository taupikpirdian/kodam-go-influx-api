# Specify — Get Available Date For Playback Personel

## 🎯 Tujuan
Menentukan kebutuhan dan kontrak fitur **Get Available Date For Playback Personel**.
Fitur ini berfokus pada pengambilan data tanggal, di tanggal berapa saja ada data di influxdb

---

## 🧩 Feature / Scope

- **Nama Bounded Context:** Get Available Date For Playback Personel
- **Ruang lingkup:** List
- **Nilai bisnis:**  
  Memudahkan user dalam melihat ada di tanggal berapa aja data yang tersedia di influx db

## 🧪 Use Cases

### 1️⃣ List Data Tanggal
API: /api/sensors/available-data/personel?date=2025-11-01T00:00:00Z&client_code=kodam

**Query Param:**
```json
{
  "date": "2025-11-01",
  "client_code": "kodam",
}
```

**Response (Success):**
```json
{
  "status": true,
  "message": "success",
  "data": [
    "2025-11-01"
    "2025-11-02"
    "2025-11-10"
    "2025-11-11"
  ]
}
```

**Aturan Cara Pengambilan Data:**
- get berdasarkan r._measurement == "personel_sensor"