# KODAM InfluxDB API Documentation

This directory contains the OpenAPI 3.0 specification for the KODAM InfluxDB API.

## Files

- `openapi.yaml` - OpenAPI 3.0 specification in YAML format
- `openapi.json` - OpenAPI 3.0 specification in JSON format

## API Overview

The KODAM InfluxDB API provides endpoints for storing and retrieving sensor data from various sources including:

- **Personel Sensors** - Personnel tracking and monitoring data
- **Radar Sensors** - Radar detection and tracking data
- **DF Sensors** - Direction Finding sensor data
- **ADSB Sensors** - Aircraft surveillance data

### Key Features

- **Real-time Streaming**: NDJSON streaming for real-time data delivery
- **Pagination**: Support for paginated results
- **Time-based Filtering**: Filter data by time ranges
- **Client-based Filtering**: Filter by client code
- **Security**: API key authentication, IP whitelisting, and rate limiting
- **Flexible Data**: JSON data storage for flexible sensor data formats

## Authentication

All API endpoints (except `/health`) require authentication using HTTP headers:

```
X-API-Key: your-api-key
X-API-Secret: your-api-secret
```

### Security Features

- **API Key Authentication**: Validates both API key and secret
- **IP Whitelist**: Restricts access by IP address
- **Rate Limiting**: 30 requests per minute per client
- **CORS**: Configurable cross-origin resource sharing

## Using the Documentation

### Swagger UI

You can view the interactive API documentation using Swagger UI:

1. **Online**: Upload the `openapi.yaml` or `openapi.json` file to [Swagger Editor](https://editor.swagger.io/)
2. **Local Setup**:
   ```bash
   # Using Docker
   docker run -p 80:8080 -e SWAGGER_JSON=/openapi.yaml -v $(pwd)/openapi.yaml:/openapi.yaml swaggerapi/swagger-ui
   ```

### Code Generation

Generate client SDKs or server stubs using the OpenAPI Generator:

```bash
# Generate Go client
docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g go \
    -o /local/generated/go

# Generate TypeScript client
docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g typescript-axios \
    -o /local/generated/typescript

# Generate Python client
docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g python \
    -o /local/generated/python
```

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Personel Sensors
- `POST /api/sensors/personel` - Store personel sensor data
- `GET /api/sensors/personel` - Retrieve personel data (NDJSON streaming)
- `GET /api/sensors/personel/list` - Retrieve personel data (JSON array with pagination)

### Radar Sensors
- `POST /api/sensors/radar` - Store radar sensor data
- `GET /api/sensors/radar` - Retrieve radar data (NDJSON streaming)
- `GET /api/sensors/radar/list` - Retrieve radar data (JSON array with pagination)

### DF Sensors
- `POST /api/sensors/df` - Store DF sensor data
- `GET /api/sensors/df` - Retrieve DF data (NDJSON streaming)
- `GET /api/sensors/df/list` - Retrieve DF data (JSON array with pagination)

### ADSB Sensors
- `POST /api/sensors/adsb` - Store ADSB sensor data
- `GET /api/sensors/adsb` - Retrieve ADSB data (NDJSON streaming)
- `GET /api/sensors/adsb/list` - Retrieve ADSB data (JSON array with pagination)

### Testing
- `POST /api/test/influx` - Test InfluxDB connection

## Data Format

### Request Format (POST endpoints)
```json
{
  "timestamp": "2025-11-08T16:31:00Z",
  "client_code": "kodam",
  "json_data": "{\"sensor_specific_field\": \"value\", ...}"
}
```

### Response Format
All API responses follow this standard format:

```json
{
  "success": true,
  "status": 200,
  "status_message": "OK",
  "message": "Operation completed successfully",
  "data": null
}
```

### List Response Format
Paginated endpoints return this format:

```json
{
  "status": "success",
  "code": 200,
  "message": "Data retrieved successfully",
  "data": [
    {
      "timestamp": "2025-11-08T16:31:00Z",
      "client_code": "kodam",
      "json_data": "{\"sensor_data\": \"...\"}"
    }
  ],
  "meta": {
    "count": 50,
    "page": 1,
    "limit": 50
  }
}
```

## Query Parameters

### Common Parameters

- `client_code` (string, optional) - Filter by client code
- `start` (datetime, optional) - Start time in RFC3339 format
- `stop` (datetime, optional) - End time in RFC3339 format
- `page` (integer, optional) - Page number for pagination (default: 1)
- `limit` (integer, optional) - Items per page (default: 50, max: 1000)

### Time Format
All time parameters should be in RFC3339 format:
```
2025-11-08T16:31:00Z
2025-11-08T23:59:59+07:00
```

## Error Handling

### HTTP Status Codes

- `200` - Success
- `400` - Bad Request (invalid parameters or data)
- `401` - Unauthorized (missing or invalid API credentials)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

### Error Response Format
```json
{
  "success": false,
  "status": 400,
  "status_message": "Bad Request",
  "message": "Invalid request parameters",
  "data": null
}
```

## Streaming Data

For real-time data retrieval, use the streaming endpoints (without `/list`):

- Response format: NDJSON (Newline Delimited JSON)
- Each line is a separate JSON object
- Continuous data stream for real-time updates

Example streaming response:
```
{"timestamp":"2025-11-08T16:31:00Z","client_code":"kodam","json_data":"{\"sensor_id\":\"S001\",...}"}
{"timestamp":"2025-11-08T16:31:30Z","client_code":"kodam","json_data":"{\"sensor_id\":\"S002\",...}"}
```

## Code Examples

### cURL
```bash
# Store personel data
curl -X POST http://localhost:3000/api/sensors/personel \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "X-API-Secret: your-api-secret" \
  -d '{
    "timestamp": "2025-11-08T16:31:00Z",
    "client_code": "kodam",
    "json_data": "{\"person_id\":\"P12345\",\"name\":\"Sersan Budi\"}"
  }'

# Retrieve personel data with pagination
curl -X GET "http://localhost:3000/api/sensors/personel/list?client_code=kodam&page=1&limit=10" \
  -H "X-API-Key: your-api-key" \
  -H "X-API-Secret: your-api-secret"
```

### Go
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func main() {
    // Store data
    data := map[string]interface{}{
        "timestamp":   "2025-11-08T16:31:00Z",
        "client_code": "kodam",
        "json_data":   "{\"person_id\":\"P12345\"}",
    }

    jsonData, _ := json.Marshal(data)

    req, _ := http.NewRequest("POST", "http://localhost:3000/api/sensors/personel", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "your-api-key")
    req.Header.Set("X-API-Secret", "your-api-secret")

    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()
}
```

### Python
```python
import requests
import json

# Store data
url = "http://localhost:3000/api/sensors/personel"
headers = {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key",
    "X-API-Secret": "your-api-secret"
}

data = {
    "timestamp": "2025-11-08T16:31:00Z",
    "client_code": "kodam",
    "json_data": json.dumps({"person_id": "P12345", "name": "Sersan Budi"})
}

response = requests.post(url, headers=headers, json=data)
print(response.json())

# Streaming data
response = requests.get(
    "http://localhost:3000/api/sensors/personel?client_code=kodam",
    headers=headers,
    stream=True
)

for line in response.iter_lines():
    if line:
        data = json.loads(line.decode('utf-8'))
        print(data)
```

## Development

### Updating the Documentation

When adding new endpoints or modifying existing ones:

1. Update the `openapi.yaml` file
2. Validate the specification:
   ```bash
   # Using swagger-codegen
   docker run --rm -v "${PWD}:/local" swaggerapi/swagger-codegen validate -i /local/openapi.yaml

   # Using openapi-generator
   docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli validate -i /local/openapi.yaml
   ```

3. Convert to JSON format:
   ```bash
   # Using yaml-to-json
   python -c "import json, yaml; print(json.dumps(yaml.safe_load(open('openapi.yaml')), indent=2))" > openapi.json
   ```

### Best Practices

1. **Version Control**: Keep the OpenAPI specification in version control
2. **Validation**: Always validate the specification after changes
3. **Examples**: Include comprehensive examples for all endpoints
4. **Documentation**: Keep descriptions clear and up-to-date
5. **Security**: Document all security requirements and authentication methods

## Support

For API support and questions:
- Email: support@kodam.mil.id
- Documentation: Available in this repository
- Issues: Create an issue in the project repository