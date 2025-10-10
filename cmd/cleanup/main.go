package main

// Scheduled Cleanup for InfluxDB: delete data older than retention threshold.
// Uses Flux Delete API via influxdb-client-go/v2.

import (
    "context"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "github.com/joho/godotenv"
    "github.com/robfig/cron/v3"
)

// parseBoolEnv returns true if the env string is one of: 1,true,yes,on.
func parseBoolEnv(v string) bool {
    switch strings.ToLower(strings.TrimSpace(v)) {
    case "1", "true", "yes", "on":
        return true
    default:
        return false
    }
}

// countOlderThan queries total rows older than cutoff for field json_data.
func countOlderThan(ctx context.Context, client influxdb2.Client, org, bucket string, clientCode string, cutoff time.Time) (int64, error) {
    flux := fmt.Sprintf(`from(bucket: %q)
        |> range(start: time(v: 0), stop: time(v: %s))
        |> filter(fn: (r) => r._field == "json_data")
        |> filter(fn: (r) => r.client_code == %q)
        |> count()
        |> keep(columns: ["_value"])`, bucket, cutoff.UTC().Format(time.RFC3339), clientCode)
    q := client.QueryAPI(org)
    res, err := q.Query(ctx, flux)
    if err != nil {
        return 0, err
    }
    var total int64
    for res.Next() {
        rec := res.Record()
        // _value is int count per series; sum all
        switch v := rec.Value().(type) {
        case int64:
            total += v
        case int:
            total += int64(v)
        case uint64:
            total += int64(v)
        case float64:
            total += int64(v)
        default:
            // attempt string parse
            s := fmt.Sprintf("%v", v)
            if n, err := strconv.ParseInt(s, 10, 64); err == nil {
                total += n
            }
        }
    }
    if res.Err() != nil {
        return 0, res.Err()
    }
    return total, nil
}

// runCleanup performs the deletion for data older than cutoff.
func runCleanup(ctx context.Context) error {
    _ = godotenv.Load()

    enabled := parseBoolEnv(os.Getenv("CLEANUP_ENABLED"))
    retentionDaysStr := strings.TrimSpace(os.Getenv("CLEANUP_RETENTION_DAYS"))
    if retentionDaysStr == "" {
        retentionDaysStr = "30"
    }
    retentionDays, err := strconv.Atoi(retentionDaysStr)
    if err != nil || retentionDays <= 0 {
        retentionDays = 30
    }

    influxURL := os.Getenv("INFLUX_URL")
    influxToken := os.Getenv("INFLUX_TOKEN")
    influxOrg := os.Getenv("INFLUX_ORG")
    influxBucket := os.Getenv("INFLUX_BUCKET")
    if influxURL == "" || influxToken == "" || influxOrg == "" || influxBucket == "" {
        return fmt.Errorf("missing Influx env: INFLUX_URL/INFLUX_TOKEN/INFLUX_ORG/INFLUX_BUCKET")
    }

    client := influxdb2.NewClient(influxURL, influxToken)
    defer client.Close()

    cutoff := time.Now().UTC().Add(-time.Duration(retentionDays) * 24 * time.Hour)
    log.Printf("[cleanup] Retention: %d days, cutoff: %s", retentionDays, cutoff.Format(time.RFC3339))

    // Optional statistics (pre-delete): count rows older than cutoff
    clientCode := strings.TrimSpace(os.Getenv("CLIENT_CODE"))
    if clientCode == "" {
        return fmt.Errorf("missing CLIENT_CODE env: set CLIENT_CODE to scope cleanup by client")
    }

    preCount, err := countOlderThan(ctx, client, influxOrg, influxBucket, clientCode, cutoff)
    if err != nil {
        log.Printf("[cleanup] pre-count error: %v", err)
    } else {
        log.Printf("[cleanup] rows older than cutoff (pre): %d", preCount)
    }

    if !enabled {
        log.Printf("[cleanup] CLEANUP_ENABLED is false. Dry-run only. No deletion performed.")
        return nil
    }

    // Execute delete for time range [1970, cutoff)
    start := time.Unix(0, 0).UTC()
    stop := cutoff
    del := client.DeleteAPI()
    // Scope deletion by client_code (and field json_data for safety)
    predicate := fmt.Sprintf("client_code=\"%s\" AND _field=\"json_data\"", clientCode)
    if err := del.DeleteWithName(ctx, influxOrg, influxBucket, start, stop, predicate); err != nil {
        return fmt.Errorf("delete API error: %w", err)
    }
    log.Printf("[cleanup] delete executed for range [%s .. %s)", start.Format(time.RFC3339), stop.Format(time.RFC3339))

    // Optional statistics (post-delete)
    postCount, err := countOlderThan(ctx, client, influxOrg, influxBucket, clientCode, cutoff)
    if err != nil {
        log.Printf("[cleanup] post-count error: %v", err)
    } else {
        log.Printf("[cleanup] rows older than cutoff (post): %d", postCount)
    }

    return nil
}

func main() {
    _ = godotenv.Load()
    cronSpec := strings.TrimSpace(os.Getenv("CLEANUP_CRON_EXPRESSION"))
    if cronSpec == "" {
        // default: run once immediately
        if err := runCleanup(context.Background()); err != nil {
            log.Fatalf("cleanup failed: %v", err)
        }
        return
    }

    log.Printf("[scheduler] Starting cleanup cron: %s", cronSpec)
    c := cron.New()
    _, err := c.AddFunc(cronSpec, func() {
        if err := runCleanup(context.Background()); err != nil {
            log.Printf("cleanup failed: %v", err)
        }
    })
    if err != nil {
        log.Fatalf("failed to start scheduler: %v", err)
    }
    c.Start()
    select {}
}