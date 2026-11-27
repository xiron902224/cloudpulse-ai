# API Reference

The CloudPulse AI API runs on `http://localhost:8080` by default.

## Health

### `GET /health`

Returns the current API health status.

Example response:

```json
{
  "status": "healthy",
  "service": "cloudpulse-api",
  "version": "0.1.0",
  "timestamp": "2026-09-12T00:00:00Z"
}
```

## Prometheus Metrics

### `GET /metrics`

Exposes application metrics in Prometheus text format. Prometheus scrapes this endpoint every 15 seconds when the Docker Compose monitoring stack is running.

The middleware records:

- `http_requests_total`: total HTTP requests by method, path, and status.
- `http_request_duration_seconds`: HTTP request duration histogram by method and path.
- `http_errors_total`: HTTP responses with status `400` or higher.

Example:

```bash
curl http://localhost:8080/metrics
```

## Metrics Data API

The metrics data endpoints are grouped under `/api/v1/metrics`.

### `POST /api/v1/metrics`

Creates a monitoring metric. `cpu_usage` and `memory_usage` must be between `0` and `100`.

### `GET /api/v1/metrics`

Lists metrics with pagination. Supported query parameters:

- `limit`: maximum number of results, capped at `100`.
- `offset`: number of results to skip.

### `GET /api/v1/metrics/:id`

Returns one metric by UUID.

### `DELETE /api/v1/metrics/:id`

Deletes one metric by UUID.

## Alerts API

Alert endpoints are grouped under `/api/v1/alerts`.

### `POST /api/v1/alerts`

Creates an alert. Supported severities are `critical`, `warning`, and `info`.

### `GET /api/v1/alerts`

Lists alerts. Optional filters are `severity` and `status`.

### `PUT /api/v1/alerts/:id`

Updates the provided fields of an existing alert.

## Interactive Documentation

Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```
