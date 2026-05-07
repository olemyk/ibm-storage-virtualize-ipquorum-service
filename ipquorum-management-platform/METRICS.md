# Prometheus Metrics Documentation

## Overview

The IP Quorum Management Platform exposes Prometheus-compatible metrics at the `/metrics` endpoint. These metrics provide comprehensive observability into the platform's operation, performance, and health.

## Accessing Metrics

**Endpoint:** `http://localhost:8080/metrics`

**Authentication:** None required (public endpoint)

**Format:** Prometheus text-based exposition format

## Available Metrics

### HTTP Metrics

#### `ipquorum_http_requests_total`
**Type:** Counter  
**Description:** Total number of HTTP requests processed  
**Labels:**
- `method`: HTTP method (GET, POST, PUT, DELETE, etc.)
- `path`: Request path (e.g., `/api/v1/instances`)
- `status`: HTTP status code (200, 404, 500, etc.)

**Example:**
```
ipquorum_http_requests_total{method="GET",path="/api/v1/instances",status="200"} 42
ipquorum_http_requests_total{method="POST",path="/api/v1/auth/login",status="200"} 5
```

#### `ipquorum_http_request_duration_seconds`
**Type:** Histogram  
**Description:** HTTP request duration in seconds  
**Labels:**
- `method`: HTTP method
- `path`: Request path

**Buckets:** 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10

**Example:**
```
ipquorum_http_request_duration_seconds_bucket{method="GET",path="/api/v1/instances",le="0.1"} 40
ipquorum_http_request_duration_seconds_sum{method="GET",path="/api/v1/instances"} 2.5
ipquorum_http_request_duration_seconds_count{method="GET",path="/api/v1/instances"} 42
```

#### `ipquorum_http_request_size_bytes`
**Type:** Summary  
**Description:** HTTP request size in bytes  
**Labels:**
- `method`: HTTP method
- `path`: Request path

#### `ipquorum_http_response_size_bytes`
**Type:** Summary  
**Description:** HTTP response size in bytes  
**Labels:**
- `method`: HTTP method
- `path`: Request path

### Instance Metrics

#### `ipquorum_instances_total`
**Type:** Gauge  
**Description:** Total number of IP Quorum instances managed by the platform

**Example:**
```
ipquorum_instances_total 5
```

#### `ipquorum_instances_healthy`
**Type:** Gauge  
**Description:** Number of healthy IP Quorum instances

#### `ipquorum_instances_unhealthy`
**Type:** Gauge  
**Description:** Number of unhealthy IP Quorum instances

#### `ipquorum_instances_degraded`
**Type:** Gauge  
**Description:** Number of degraded IP Quorum instances

#### `ipquorum_instance_operations_total`
**Type:** Counter  
**Description:** Total number of instance operations performed  
**Labels:**
- `operation`: Operation type (create, start, stop, restart, delete)
- `status`: Operation status (success, failure)

**Example:**
```
ipquorum_instance_operations_total{operation="start",status="success"} 15
ipquorum_instance_operations_total{operation="start",status="failure"} 2
```

#### `ipquorum_instance_operation_duration_seconds`
**Type:** Histogram  
**Description:** Instance operation duration in seconds  
**Labels:**
- `operation`: Operation type

**Buckets:** 0.1, 0.5, 1, 2, 5, 10, 30, 60

### Authentication Metrics

#### `ipquorum_auth_attempts_total`
**Type:** Counter  
**Description:** Total number of authentication attempts  
**Labels:**
- `status`: Attempt status (success, failure)

**Example:**
```
ipquorum_auth_attempts_total{status="success"} 100
ipquorum_auth_attempts_total{status="failure"} 5
```

#### `ipquorum_auth_tokens_issued_total`
**Type:** Counter  
**Description:** Total number of authentication tokens issued

#### `ipquorum_auth_tokens_active`
**Type:** Gauge  
**Description:** Number of currently active authentication tokens

### Database Metrics

#### `ipquorum_db_queries_total`
**Type:** Counter  
**Description:** Total number of database queries executed  
**Labels:**
- `operation`: Query operation (select, insert, update, delete)
- `status`: Query status (success, failure)

#### `ipquorum_db_query_duration_seconds`
**Type:** Histogram  
**Description:** Database query duration in seconds  
**Labels:**
- `operation`: Query operation

**Buckets:** 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1

#### `ipquorum_db_connections_active`
**Type:** Gauge  
**Description:** Number of active database connections

### Health Check Metrics

#### `ipquorum_health_checks_total`
**Type:** Counter  
**Description:** Total number of health checks performed  
**Labels:**
- `instance_id`: Instance UUID
- `status`: Health check result (healthy, unhealthy, degraded)

#### `ipquorum_health_check_duration_seconds`
**Type:** Histogram  
**Description:** Health check duration in seconds  
**Labels:**
- `instance_id`: Instance UUID

**Buckets:** 0.5, 1, 2, 5, 10, 30

### Script Execution Metrics

#### `ipquorum_script_executions_total`
**Type:** Counter  
**Description:** Total number of bash script executions  
**Labels:**
- `script`: Script name (e.g., ipquorum-instance-manager.sh)
- `status`: Execution status (success, failure)

#### `ipquorum_script_execution_duration_seconds`
**Type:** Histogram  
**Description:** Script execution duration in seconds  
**Labels:**
- `script`: Script name

**Buckets:** 0.1, 0.5, 1, 2, 5, 10, 30, 60, 120

## Prometheus Configuration

### Scrape Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'ipquorum-management'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 15s
    scrape_timeout: 10s
    metrics_path: '/metrics'
```

### Example Queries

#### Request Rate
```promql
rate(ipquorum_http_requests_total[5m])
```

#### Error Rate
```promql
rate(ipquorum_http_requests_total{status=~"5.."}[5m])
```

#### Average Request Duration
```promql
rate(ipquorum_http_request_duration_seconds_sum[5m]) / 
rate(ipquorum_http_request_duration_seconds_count[5m])
```

#### Instance Health Status
```promql
ipquorum_instances_healthy / ipquorum_instances_total
```

#### Authentication Success Rate
```promql
rate(ipquorum_auth_attempts_total{status="success"}[5m]) /
rate(ipquorum_auth_attempts_total[5m])
```

#### P95 Request Latency
```promql
histogram_quantile(0.95, 
  rate(ipquorum_http_request_duration_seconds_bucket[5m])
)
```

## Grafana Dashboard

### Recommended Panels

1. **Request Rate** - Graph showing requests per second
2. **Error Rate** - Graph showing 4xx and 5xx errors
3. **Request Duration** - Heatmap of request latencies
4. **Instance Health** - Gauge showing healthy/unhealthy instances
5. **Authentication Activity** - Counter of login attempts
6. **Database Performance** - Query duration histogram
7. **Active Connections** - Gauge of active DB connections

### Sample Dashboard JSON

A complete Grafana dashboard configuration will be provided in a future update.

## Alerting Rules

### Example Prometheus Alert Rules

```yaml
groups:
  - name: ipquorum_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(ipquorum_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} requests/sec"

      - alert: InstanceUnhealthy
        expr: ipquorum_instances_unhealthy > 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Unhealthy instances detected"
          description: "{{ $value }} instances are unhealthy"

      - alert: SlowRequests
        expr: histogram_quantile(0.95, rate(ipquorum_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow API requests"
          description: "P95 latency is {{ $value }}s"

      - alert: AuthenticationFailures
        expr: rate(ipquorum_auth_attempts_total{status="failure"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High authentication failure rate"
          description: "{{ $value }} failed auth attempts per second"
```

## Best Practices

1. **Scrape Interval**: Use 15-30 second intervals for production
2. **Retention**: Keep metrics for at least 15 days
3. **Cardinality**: Monitor label cardinality to avoid explosion
4. **Aggregation**: Use recording rules for frequently queried metrics
5. **Alerting**: Set up alerts for critical metrics (error rate, instance health)

## Troubleshooting

### Metrics Not Appearing

1. Check that the server is running: `curl http://localhost:8080/health`
2. Verify metrics endpoint: `curl http://localhost:8080/metrics`
3. Check Prometheus scrape status in Prometheus UI

### High Cardinality

If you see performance issues:
1. Review label values (especially dynamic ones like instance_id)
2. Consider aggregating metrics before export
3. Use recording rules for complex queries

### Missing Data

1. Verify scrape interval matches your query range
2. Check for network issues between Prometheus and the platform
3. Review Prometheus logs for scrape errors

## Future Enhancements

- [ ] Custom business metrics (quorum operations, data transfer)
- [ ] Distributed tracing integration (OpenTelemetry)
- [ ] Metrics aggregation across multiple servers
- [ ] Real-time metrics streaming via WebSocket
- [ ] Metrics export to multiple backends (InfluxDB, Datadog)