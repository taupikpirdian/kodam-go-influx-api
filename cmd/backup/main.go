package main

// Data Backup: Query InfluxDB (last 24h) -> CSV -> gzip -> store (local/MinIO/both)

import (
    "compress/gzip"
    "context"
    "encoding/csv"
    "fmt"
    "io"
    "log"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "time"

    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "github.com/joho/godotenv"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "github.com/robfig/cron/v3"
)

// queryToCSV streams results from InfluxDB into a CSV file with header: time,user_id,json
// Only rows with tag client_code == clientCode will be included.
func queryToCSV(ctx context.Context, influxURL, token, org, bucket string, clientCode string, start, stop time.Time, outPath string) (int, error) {
    client := influxdb2.NewClient(influxURL, token)
    defer client.Close()

    q := client.QueryAPI(org)
    flux := fmt.Sprintf(`from(bucket: %q)
        |> range(start: time(v: %s), stop: time(v: %s))
        |> filter(fn: (r) => r._field == "json_data")
        |> filter(fn: (r) => r.client_code == %q)
        |> keep(columns: ["_time", "client_code", "_value"]) 
        |> sort(columns: ["_time"], desc: false)`, bucket, start.UTC().Format(time.RFC3339), stop.UTC().Format(time.RFC3339), clientCode)

    res, err := q.Query(ctx, flux)
    if err != nil {
        return 0, err
    }

    f, err := os.Create(outPath)
    if err != nil {
        return 0, err
    }
    defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    // Write header
    if err := w.Write([]string{"time", "user_id", "json"}); err != nil {
        return 0, err
    }

    count := 0
    for res.Next() {
        rec := res.Record()
        ts := rec.Time().UTC().Format(time.RFC3339)

        // client_code is a tag
        var clientCode string
        if v := rec.ValueByKey("client_code"); v != nil {
            if s, ok := v.(string); ok {
                clientCode = s
            } else {
                clientCode = fmt.Sprintf("%v", v)
            }
        }

        // _value is json_data field
        var jsonData string
        if v := rec.Value(); v != nil {
            if s, ok := v.(string); ok {
                jsonData = s
            } else {
                jsonData = fmt.Sprintf("%v", v)
            }
        }

        if err := w.Write([]string{ts, clientCode, jsonData}); err != nil {
            return 0, err
        }
        count++
    }
    if res.Err() != nil {
        return 0, res.Err()
    }
    return count, nil
}

// gzipFile compresses the input file to output .gz
func gzipFile(inPath, outPath string) error {
    in, err := os.Open(inPath)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(outPath)
    if err != nil {
        return err
    }
    defer out.Close()

    gw := gzip.NewWriter(out)
    gw.Name = filepath.Base(inPath)
    gw.ModTime = time.Now()
    defer gw.Close()

    if _, err := io.Copy(gw, in); err != nil {
        return err
    }
    return nil
}

// saveToLocal stores gz file under LOCAL_BACKUP_DIR/YYYY/MM/backup-YYYYMM.csv.gz
// If called multiple times within the same year/month, the file will be replaced.
func saveToLocal(srcGzPath, baseDir string, now time.Time, tsStr string) (string, error) {
    // Ensure baseDir exists
    if baseDir == "" {
        baseDir = "./backups"
    }
    // Build nested dir and file name (monthly)
    localDir := filepath.Join(baseDir, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", int(now.Month())))
    if err := os.MkdirAll(localDir, 0o755); err != nil {
        return "", err
    }
    // Single file per month (replaces when exists)
    dstPath := filepath.Join(localDir, fmt.Sprintf("backup-%04d%02d.csv.gz", now.Year(), int(now.Month())))

    // Copy file
    in, err := os.Open(srcGzPath)
    if err != nil {
        return "", err
    }
    defer in.Close()
    out, err := os.Create(dstPath)
    if err != nil {
        return "", err
    }
    defer out.Close()
    if _, err := io.Copy(out, in); err != nil {
        return "", err
    }
    return dstPath, nil
}

// uploadToMinio uploads the file to MinIO and returns an accessible URL
func uploadToMinio(ctx context.Context, endpoint, accessKey, secretKey, bucket, objectName, filePath string) (string, error) {
    // Parse endpoint to determine scheme and host
    var host string
    var secure bool
    if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
        host = u.Host
        secure = strings.EqualFold(u.Scheme, "https")
    } else {
        host = endpoint
        secure = false
    }

    cl, err := minio.New(host, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: secure})
    if err != nil {
        return "", err
    }

    // Ensure bucket exists
    exists, err := cl.BucketExists(ctx, bucket)
    if err != nil {
        return "", err
    }
    if !exists {
        if err := cl.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
            return "", err
        }
    }

    // Upload
    _, err = cl.FPutObject(ctx, bucket, objectName, filePath, minio.PutObjectOptions{ContentType: "application/gzip"})
    if err != nil {
        return "", err
    }

    // Construct public URL (path-style)
    base := endpoint
    if u, err := url.Parse(endpoint); err == nil && u.Host != "" && u.Scheme != "" {
        base = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
    }
    urlStr := fmt.Sprintf("%s/%s/%s", strings.TrimRight(base, "/"), bucket, objectName)
    return urlStr, nil
}

// runBackup orchestrates the backup: query -> CSV -> gzip -> upload
func runBackup(ctx context.Context) error {
    // Load env
    _ = godotenv.Load()

    influxURL := os.Getenv("INFLUX_URL")
    influxToken := os.Getenv("INFLUX_TOKEN")
    influxOrg := os.Getenv("INFLUX_ORG")
    influxBucket := os.Getenv("INFLUX_BUCKET")

    // Destination mode
    storageMode := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_MODE")))
    if storageMode == "" {
        storageMode = "local" // default to local to avoid external deps by default
    }
    localBackupDir := strings.TrimSpace(os.Getenv("LOCAL_BACKUP_DIR"))
    if localBackupDir == "" {
        localBackupDir = "./backups"
    }

    // MinIO settings (only required if storageMode involves minio)
    minioEndpoint := os.Getenv("MINIO_ENDPOINT")
    minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
    minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
    minioBucket := os.Getenv("MINIO_BUCKET")
    clientCode := strings.TrimSpace(os.Getenv("CLIENT_CODE"))

    if influxURL == "" || influxToken == "" || influxOrg == "" || influxBucket == "" {
        return fmt.Errorf("missing Influx env: INFLUX_URL/INFLUX_TOKEN/INFLUX_ORG/INFLUX_BUCKET")
    }
    // Validate destination-specific env
    switch storageMode {
    case "local":
        // no additional validation
    case "minio":
        if minioEndpoint == "" || minioAccessKey == "" || minioSecretKey == "" || minioBucket == "" {
            return fmt.Errorf("missing MinIO env for storage mode 'minio': MINIO_ENDPOINT/MINIO_ACCESS_KEY/MINIO_SECRET_KEY/MINIO_BUCKET")
        }
    case "both":
        if minioEndpoint == "" || minioAccessKey == "" || minioSecretKey == "" || minioBucket == "" {
            return fmt.Errorf("missing MinIO env for storage mode 'both': MINIO_ENDPOINT/MINIO_ACCESS_KEY/MINIO_SECRET_KEY/MINIO_BUCKET")
        }
    default:
        return fmt.Errorf("invalid STORAGE_MODE: %s (use local|minio|both)", storageMode)
    }
    if clientCode == "" {
        return fmt.Errorf("missing CLIENT_CODE env: set CLIENT_CODE to filter backup by client")
    }

    // Time range: last 24 hours
    stop := time.Now().UTC()
    start := stop.Add(-24 * time.Hour)

    // Local temp file paths
    tsStr := time.Now().UTC().Format("20060102T150405Z")
    tmpDir := os.TempDir()
    csvPath := filepath.Join(tmpDir, fmt.Sprintf("backup-%s.csv", tsStr))
    gzPath := csvPath + ".gz"

    log.Printf("[backup] Querying InfluxDB from %s to %s...", start.Format(time.RFC3339), stop.Format(time.RFC3339))
    rows, err := queryToCSV(ctx, influxURL, influxToken, influxOrg, influxBucket, clientCode, start, stop, csvPath)
    if err != nil {
        return fmt.Errorf("queryToCSV error: %w", err)
    }
    log.Printf("[backup] Wrote %d rows to %s", rows, csvPath)

    log.Printf("[backup] Compressing to %s...", gzPath)
    if err := gzipFile(csvPath, gzPath); err != nil {
        return fmt.Errorf("gzipFile error: %w", err)
    }

    now := time.Now().UTC()
    // Local and/or MinIO persistence
    switch storageMode {
    case "local":
        dst, err := saveToLocal(gzPath, localBackupDir, now, tsStr)
        if err != nil {
            return fmt.Errorf("saveToLocal error: %w", err)
        }
        log.Printf("[backup] Saved locally: %s", dst)
        fmt.Println(dst)
    case "minio":
        // Single file per month in MinIO as well
        objectName := fmt.Sprintf("backups/%04d/%02d/backup-%04d%02d.csv.gz", now.Year(), now.Month(), now.Year(), now.Month())
        log.Printf("[backup] Uploading to MinIO bucket %s as %s...", minioBucket, objectName)
        fileURL, err := uploadToMinio(ctx, minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, objectName, gzPath)
        if err != nil {
            return fmt.Errorf("uploadToMinio error: %w", err)
        }
        log.Printf("[backup] Upload success: %s", fileURL)
        fmt.Println(fileURL)
    case "both":
        // Local save
        dst, err := saveToLocal(gzPath, localBackupDir, now, tsStr)
        if err != nil {
            return fmt.Errorf("saveToLocal error: %w", err)
        }
        log.Printf("[backup] Saved locally: %s", dst)
        // MinIO upload (monthly)
        objectName := fmt.Sprintf("backups/%04d/%02d/backup-%04d%02d.csv.gz", now.Year(), now.Month(), now.Year(), now.Month())
        log.Printf("[backup] Uploading to MinIO bucket %s as %s...", minioBucket, objectName)
        fileURL, err := uploadToMinio(ctx, minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, objectName, gzPath)
        if err != nil {
            return fmt.Errorf("uploadToMinio error: %w", err)
        }
        log.Printf("[backup] Upload success: %s", fileURL)
        fmt.Println(fileURL)
    }
    return nil
}

func main() {
    // If SCHEDULE_CRON is set, run with robfig/cron; else run once.
    _ = godotenv.Load()
    cronSpec := os.Getenv("SCHEDULE_CRON")
    if strings.TrimSpace(cronSpec) == "" {
        if err := runBackup(context.Background()); err != nil {
            log.Fatalf("backup failed: %v", err)
        }
        return
    }

    log.Printf("[scheduler] Starting cron with spec: %s", cronSpec)
    c := cron.New()
    _, err := c.AddFunc(cronSpec, func() {
        if err := runBackup(context.Background()); err != nil {
            log.Printf("backup failed: %v", err)
        }
    })
    if err != nil {
        log.Fatalf("failed to start scheduler: %v", err)
    }
    c.Start()
    // Block forever
    select {}
}