package main

// Seed script: menulis ribuan data dummy ke InfluxDB

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "os"
    "time"

    "rti/influxdb/domain"
    "rti/influxdb/infrastructure/config"
    "rti/influxdb/infrastructure/repository"
)

type DummyPayload struct {
    PersonID   string  `json:"person_id"`
    Name       string  `json:"name"`
    Location   struct {
        Latitude  float64 `json:"latitude"`
        Longitude float64 `json:"longitude"`
    } `json:"location"`
    Status     string  `json:"status"`
    HeartRate  int     `json:"heart_rate"`
    Temperature float64 `json:"temperature"`
}

func main() {
    // Flags dengan default dari env (jika ada)
    defaultCount := getenvInt("SEED_COUNT", 5000)
    defaultClient := getenv("SEED_CLIENT", "seed")
    defaultInterval := getenvDuration("SEED_INTERVAL", time.Second)
    defaultStart := getenv("SEED_START", "")

    var (
        count    = flag.Int("count", defaultCount, "jumlah data dummy yang akan ditulis")
        client   = flag.String("client", defaultClient, "nilai client_code untuk tag InfluxDB")
        interval = flag.Duration("interval", defaultInterval, "interval antar timestamp setiap record")
        startStr = flag.String("start", defaultStart, "waktu mulai (RFC3339), default now jika kosong")
    )
    flag.Parse()

    // Load config (.env)
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // Inisialisasi repository
    repo := repository.NewInfluxRepository(cfg)
    defer repo.Close()

    // Cek koneksi
    if err := repo.CheckConnection(context.Background()); err != nil {
        log.Fatalf("failed to connect to InfluxDB: %v", err)
    }

    // Tentukan start time
    startTime := time.Now().UTC()
    if *startStr != "" {
        st, err := time.Parse(time.RFC3339, *startStr)
        if err != nil {
            log.Fatalf("invalid start time: %v", err)
        }
        startTime = st.UTC()
    }

    // Random seed
    rand.Seed(time.Now().UnixNano())
    names := []string{"Sersan Budi", "Kopral Andi", "Letnan Sari", "Kapten Rina", "Praka Joko"}
    statuses := []string{"active", "idle", "rest", "moving"}

    log.Printf("Starting seed: count=%d client=%s start=%s interval=%s", *count, *client, startTime.Format(time.RFC3339), interval.String())

    // Tulis data sequential
    for i := 0; i < *count; i++ {
        ts := startTime.Add(time.Duration(i) * *interval)
        payload := DummyPayload{
            PersonID:   fmt.Sprintf("P%06d", i+1),
            Name:       names[rand.Intn(len(names))],
            Status:     statuses[rand.Intn(len(statuses))],
            HeartRate:  60 + rand.Intn(60),
            Temperature: 35.5 + rand.Float64()*2.0,
        }
        payload.Location.Latitude = -6.9 + rand.Float64()*0.1
        payload.Location.Longitude = 107.6 + rand.Float64()*0.1

        b, _ := json.Marshal(payload)
        input := domain.SensorInput{
            Timestamp:  ts,
            ClientCode: *client,
            JSONData:   string(b),
        }

        if err := repo.WritePersonelSensor(context.Background(), input); err != nil {
            log.Fatalf("write failed at i=%d: %v", i, err)
        }

        if (i+1)%1000 == 0 {
            log.Printf("Written %d records", i+1)
        }
    }

    log.Printf("Seed finished: total=%d", *count)
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func getenvInt(key string, def int) int {
    if v := os.Getenv(key); v != "" {
        var n int
        if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
            return n
        }
    }
    return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
    if v := os.Getenv(key); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            return d
        }
    }
    return def
}