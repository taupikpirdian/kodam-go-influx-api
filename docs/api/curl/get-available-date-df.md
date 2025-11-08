# Get Available Date For Playback DF API

## Endpoint
```
GET /api/sensors/available-data/df
```

## Deskripsi
Mendapatkan daftar tanggal yang tersedia untuk data DF (Direction Finding) sensor di InfluxDB dalam satu bulan berdasarkan parameter tanggal yang diberikan.

## Query Parameters
- `date` (required, string): Tanggal dalam format YYYY-MM-DD untuk menentukan bulan yang akan dicari datanya
- `client_code` (required, string): Kode client untuk filter data

## Response Format

### Success (200 OK)
```json
{
  "status": true,
  "message": "success",
  "data": [
    "2025-11-01",
    "2025-11-02",
    "2025-11-10",
    "2025-11-11"
  ]
}
```

### Error (400 Bad Request)
```json
{
  "status": false,
  "message": "date parameter is required (format: YYYY-MM-DD)",
  "data": []
}
```

```json
{
  "status": false,
  "message": "client_code parameter is required",
  "data": []
}
```

```json
{
  "status": false,
  "message": "invalid date format, use YYYY-MM-DD",
  "data": []
}
```

### Error (500 Internal Server Error)
```json
{
  "status": false,
  "message": "database connection error",
  "data": []
}
```

## Contoh Penggunaan dengan cURL

### 1. Mendapatkan tanggal yang tersedia untuk November 2025
```bash
curl --location 'http://localhost:3000/api/sensors/available-data/df?date=2025-11-01&client_code=kodam'
```

### 2. Mendapatkan tanggal yang tersedia untuk Desember 2024
```bash
curl --location 'http://localhost:3000/api/sensors/available-data/df?date=2024-12-15&client_code=kodam'
```

### 3. Mendapatkan tanggal yang tersedia untuk client_code berbeda
```bash
curl --location 'http://localhost:3000/api/sensors/available-data/df?date=2025-01-10&client_code=signal_intel'
```

### 4. Error handling - tanpa parameter date
```bash
curl --location 'http://localhost:3000/api/sensors/available-data/df?client_code=kodam'
```
Response:
```json
{
  "status": false,
  "message": "date parameter is required (format: YYYY-MM-DD)",
  "data": []
}
```

### 5. Error handling - tanpa parameter client_code
```bash
curl --location 'http://localhost:3000/api/sensors/available-data/df?date=2025-11-01'
```
Response:
```json
{
  "status": false,
  "message": "client_code parameter is required",
  "data": []
}
```

### 6. Error handling - format tanggal salah
```bash
curl --location 'http://localhost:3000/api/sensors/available-data/df?date=2025/11/01&client_code=kodam'
```
Response:
```json
{
  "status": false,
  "message": "invalid date format, use YYYY-MM-DD",
  "data": []
}
```

## Cara Kerja
1. API menerima parameter `date` (YYYY-MM-DD) dan `client_code`
2. Sistem akan menentukan awal dan akhir bulan dari tanggal yang diberikan
3. Query ke InfluxDB dengan measurement `df_sensor` berdasarkan `client_code`
4. Mengembalikan daftar tanggal unik dalam format YYYY-MM-DD yang memiliki data
5. Tanggal diurutkan secara ascending (dari tanggal terlama ke terbaru)

## Aturan Cara Pengambilan Data
- Query berdasarkan `r._measurement == "df_sensor"`
- Filter berdasarkan `client_code` yang diberikan
- Menggunakan field `json_data` untuk menentukan keberadaan data
- Data DF biasanya berisi informasi seperti:
  - Signal direction (azimuth, elevation)
  - Signal strength
  - Frequency information
  - Source location estimation
  - Time of detection

## Catatan
- Query akan mencari data dalam satu bulan penuh berdasarkan tanggal yang diberikan
- Misalnya, jika `date=2025-11-01`, maka sistem akan mencari data dari 2025-11-01 sampai 2025-11-30
- Jika tidak ada data dalam bulan tersebut, akan mengembalikan array kosong
- Response mengikuti format yang telah ditentukan dalam speckit documents
- Endpoint ini secara spesifik hanya mencari data dari measurement `df_sensor`
- DF data biasanya digunakan untuk sistem intelijen sinyal dan electronic warfare

## Use Cases
- Mengecek ketersediaan data DF untuk analisis sinyal
- Validasi data untuk sistem electronic warfare
- Melihat histori data DF untuk tracking sumber sinyal
- Planning kapasitas penyimpanan data DF
- Monitoring coverage sistem deteksi arah sinyal

## DF Sensor Information
DF (Direction Finding) sensors digunakan untuk:
- Menentukan arah datangnya sinyal radio
- Melokalisasi posisi pemancar
- Monitoring spektrum frekuensi
- Identifikasi sinyal musuh dalam konteks militer
- Analisis traffic komunikasi radio