# IPQuorum Management Platform - Monitoring & Alerting

This document describes the monitoring and alerting setup for the IPQuorum Management Platform.

## Overview

The platform uses a comprehensive monitoring stack:
- **Prometheus** - Metrics collection and storage
- **Grafana** - Visualization and dashboards
- **Alertmanager** - Alert routing and notification

## Architecture

```
┌─────────────────┐
│  IPQuorum API   │──┐
│   (Go Server)   │  │ /metrics endpoint
└─────────────────┘  │
                     │
┌─────────────────┐  │
│  Web Dashboard  │  │
│   (React/TS)    │  │
└─────────────────┘  │
                     ▼
              ┌──────────────┐
              │  Prometheus  │──────┐
              │   :9090      │      │
              └──────────────┘      │
                     │              │
                     │ scrape       │ alerts
                     │              │
              ┌──────────────┐      │
              │   Grafana    │      │
              │   :3001      │      │
              └──────────────┘      │
                                    ▼
                             ┌──────────────┐
                             │ Alertmanager │
                             │   :9093      │
                             └──────────────┘
                                    │
                                    ▼
                          ┌──────────────────┐
                          │  Notifications   │
                          │ (Email/Slack/etc)│
                          └──────────────────┘
```

## Metrics

### Instance Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `ipquorum_instances_total` | Gauge | Total number of IP Quorum instances |
| `ipquorum_instances_active` | Gauge | Number of active instances |
| `ipquorum_instance_status` | Gauge | Per-instance status (0=stopped, 1=running, 2=error) |

### API Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `ipquorum_api_requests_total` | Counter | Total API requests by method, path, status |
| `ipquorum_api_request_duration_seconds` | Histogram | API request duration |
| `ipquorum_api_requests_in_flight` | Gauge | Current in-flight requests |

### System Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `go_goroutines` | Gauge | Number of goroutines |
| `go_memstats_alloc_bytes` | Gauge | Allocated memory |
| `process_cpu_seconds_total` | Counter | CPU time used |
| `process_resident_memory_bytes` | Gauge | Resident memory size |

### Database Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `ipquorum_db_connections_open` | Gauge | Open database connections |
| `ipquorum_db_connections_in_use` | Gauge | In-use database connections |
| `ipquorum_db_errors_total` | Counter | Database errors |

## Dashboards

### IPQuorum Overview Dashboard

Location: `grafana/dashboards/ipquorum-overview.json`

**Panels:**
1. **Total Instances** - Gauge showing total instance count
2. **Active Instances** - Gauge showing active instances
3. **API Request Rate** - Time series of requests/second
4. **API Response Time** - p95 and p99 latency
5. **HTTP Status Codes** - Stacked area chart of status codes
6. **Instance Status** - Table of all instances with status

**Access:** http://localhost:3001/d/ipquorum-overview

## Alerts

### Alert Rules

Location: `prometheus/alerts/ipquorum-alerts.yml`

#### Critical Alerts

| Alert | Condition | Duration | Description |
|-------|-----------|----------|-------------|
| `IPQuorumInstanceDown` | `status == 0` | 2m | Instance is down |
| `IPQuorumNoActiveInstances` | `active == 0 && total > 0` | 5m | All instances inactive |
| `IPQuorumVeryHighAPILatency` | `p95 > 5s` | 2m | Severe API slowness |
| `IPQuorumVeryHighErrorRate` | `error_rate > 10%` | 2m | High error rate |
| `IPQuorumDatabaseConnectionFailure` | `db_errors > 10/5m` | 2m | Database issues |
| `IPQuorumHealthCheckFailing` | `up == 0` | 1m | Server is down |
| `IPQuorumSuspiciousAuthActivity` | `failed_logins > 5/s` | 1m | Possible attack |

#### Warning Alerts

| Alert | Condition | Duration | Description |
|-------|-----------|----------|-------------|
| `IPQuorumInstanceError` | `status == 2` | 1m | Instance in error state |
| `IPQuorumHighAPILatency` | `p95 > 1s` | 5m | API slowness |
| `IPQuorumHighErrorRate` | `error_rate > 5%` | 5m | Elevated errors |
| `IPQuorumHighMemoryUsage` | `memory > 1GB` | 10m | High memory usage |
| `IPQuorumHighGoroutineCount` | `goroutines > 1000` | 10m | Goroutine leak |
| `IPQuorumHighAuthFailureRate` | `failed_logins > 0.5/s` | 5m | Auth failures |

#### Info Alerts

| Alert | Condition | Duration | Description |
|-------|-----------|----------|-------------|
| `IPQuorumHighRequestRate` | `requests > 100/s` | 10m | High traffic |

### Recording Rules

Pre-calculated metrics for performance:

```promql
# API request rate (5m average)
ipquorum:api_request_rate:5m

# API error rate (5m average)
ipquorum:api_error_rate:5m

# API latency p95 (5m)
ipquorum:api_latency_p95:5m

# API latency p99 (5m)
ipquorum:api_latency_p99:5m

# Instance health ratio
ipquorum:instances_health_ratio
```

## Alert Routing

### Severity Levels

- **Critical** - Immediate action required, page on-call
- **Warning** - Investigate soon, notify team
- **Info** - Informational, log only

### Notification Channels

Configure in `prometheus/alertmanager.yml`:

```yaml
receivers:
  - name: 'team-email'
    email_configs:
      - to: 'team@example.com'
        
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/...'
        channel: '#alerts'
        
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'your-key'
```

## Querying Metrics

### Prometheus UI

Access: http://localhost:9090

**Example Queries:**

```promql
# Current active instances
ipquorum_instances_active

# API request rate (last 5 minutes)
rate(ipquorum_api_requests_total[5m])

# API p95 latency
histogram_quantile(0.95, rate(ipquorum_api_request_duration_seconds_bucket[5m]))

# Error rate by endpoint
rate(ipquorum_api_requests_total{status=~"5.."}[5m])

# Memory usage trend
process_resident_memory_bytes / 1024 / 1024
```

### Grafana Explore

Access: http://localhost:3001/explore

Use for ad-hoc queries and troubleshooting.

## Best Practices

### 1. Alert Fatigue Prevention
- Set appropriate thresholds
- Use `for` duration to avoid flapping
- Group related alerts
- Use severity levels correctly

### 2. Dashboard Design
- Keep dashboards focused
- Use consistent time ranges
- Add helpful annotations
- Include links to runbooks

### 3. Metric Naming
- Follow Prometheus conventions
- Use consistent labels
- Document custom metrics
- Avoid high cardinality

### 4. Performance
- Use recording rules for complex queries
- Set appropriate scrape intervals
- Monitor Prometheus itself
- Archive old data

## Troubleshooting

### No Data in Grafana

1. Check Prometheus targets: http://localhost:9090/targets
2. Verify metrics endpoint: http://localhost:8080/metrics
3. Check Grafana datasource connection
4. Review Prometheus logs

### Alerts Not Firing

1. Check alert rules in Prometheus UI
2. Verify Alertmanager is running
3. Check alert routing configuration
4. Review Alertmanager logs

### High Memory Usage

1. Check metric cardinality
2. Review retention settings
3. Consider using recording rules
4. Archive old data

### Slow Queries

1. Use recording rules
2. Reduce time range
3. Optimize PromQL queries
4. Check Prometheus performance

## Maintenance

### Regular Tasks

- **Daily**: Review critical alerts
- **Weekly**: Check dashboard accuracy
- **Monthly**: Review alert thresholds
- **Quarterly**: Archive old metrics

### Backup

Backup Prometheus data:
```bash
# Stop Prometheus
docker-compose stop prometheus

# Backup data directory
tar -czf prometheus-backup-$(date +%Y%m%d).tar.gz prometheus/data/

# Restart Prometheus
docker-compose start prometheus
```

### Upgrades

1. Backup current configuration
2. Test in staging environment
3. Update docker-compose.yml versions
4. Apply changes: `docker-compose up -d`
5. Verify metrics collection
6. Check dashboards and alerts

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Alert Best Practices](https://prometheus.io/docs/practices/alerting/)

## Support

For issues or questions:
- Check logs: `docker-compose logs prometheus grafana`
- Review metrics: http://localhost:9090
- Consult documentation
- Contact platform team

#