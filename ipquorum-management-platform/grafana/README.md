# IPQuorum Grafana Dashboards

This directory contains Grafana dashboards and provisioning configurations for monitoring the IPQuorum Management Platform.

## Directory Structure

```
grafana/
├── dashboards/
│   └── ipquorum-overview.json       # Main overview dashboard
├── provisioning/
│   ├── dashboards/
│   │   └── dashboard.yml            # Dashboard provisioning config
│   └── datasources/
│       └── prometheus.yml           # Prometheus datasource config
└── README.md                        # This file
```

## Dashboards

### IPQuorum Overview Dashboard

The main dashboard (`ipquorum-overview.json`) provides comprehensive monitoring of the IPQuorum platform:

**Panels:**
1. **Total Instances** (Gauge) - Shows total number of IP Quorum instances
2. **Active Instances** (Gauge) - Shows currently active instances
3. **API Request Rate** (Time Series) - Requests per second by endpoint
4. **API Response Time** (Time Series) - p95 and p99 latency metrics
5. **HTTP Status Codes** (Stacked Time Series) - Status code distribution
6. **Instance Status** (Table) - Detailed instance status table

**Metrics Used:**
- `ipquorum_instances_total` - Total instance count
- `ipquorum_instances_active` - Active instance count
- `ipquorum_api_requests_total` - API request counter
- `ipquorum_api_request_duration_seconds` - Request duration histogram
- `ipquorum_instance_status` - Per-instance status

## Setup

### Using Docker Compose

The dashboards are automatically provisioned when using the provided `docker-compose.yml`:

```bash
cd ipquorum-management-platform
docker-compose up -d
```

Grafana will be available at: http://localhost:3001
- Default credentials: admin/admin (change on first login)

### Manual Setup

1. **Install Grafana**
   ```bash
   # Using Docker
   docker run -d \
     -p 3001:3000 \
     -v $(pwd)/grafana/provisioning:/etc/grafana/provisioning \
     -v $(pwd)/grafana/dashboards:/etc/grafana/provisioning/dashboards \
     --name grafana \
     grafana/grafana:latest
   ```

2. **Configure Prometheus Datasource**
   - The datasource is auto-provisioned from `provisioning/datasources/prometheus.yml`
   - Default URL: http://prometheus:9090
   - Adjust if your Prometheus is at a different location

3. **Import Dashboards**
   - Dashboards are auto-imported from `dashboards/` directory
   - Or manually import via Grafana UI: Dashboards → Import → Upload JSON

## Dashboard Features

### Auto-Refresh
- Default refresh interval: 10 seconds
- Configurable in dashboard settings

### Time Range
- Default: Last 1 hour
- Adjustable via time picker

### Variables
- Currently no template variables
- Can be added for multi-instance filtering

## Customization

### Adding New Panels

1. Edit the dashboard JSON file
2. Add new panel configuration
3. Reload Grafana or re-import dashboard

### Creating New Dashboards

1. Create dashboard in Grafana UI
2. Export as JSON
3. Save to `dashboards/` directory
4. Add to provisioning config if needed

## Metrics Reference

### Instance Metrics
```promql
# Total instances
ipquorum_instances_total

# Active instances
ipquorum_instances_active

# Instance status (0=stopped, 1=running, 2=error)
ipquorum_instance_status{instance_id="...", instance_name="..."}
```

### API Metrics
```promql
# Request rate
rate(ipquorum_api_requests_total[5m])

# Request duration (p95)
histogram_quantile(0.95, rate(ipquorum_api_request_duration_seconds_bucket[5m]))

# Error rate
rate(ipquorum_api_requests_total{status=~"5.."}[5m])
```

### System Metrics
```promql
# Go runtime metrics
go_goroutines
go_memstats_alloc_bytes
process_cpu_seconds_total
```

## Alerting

Alerts can be configured in Grafana for:
- High error rates
- Slow response times
- Instance failures
- Resource exhaustion

See the Prometheus alerting rules in `../prometheus/alerts/` for pre-configured alerts.

## Troubleshooting

### Dashboard Not Loading
- Check Grafana logs: `docker logs grafana`
- Verify provisioning directory is mounted correctly
- Ensure JSON syntax is valid

### No Data in Panels
- Verify Prometheus is scraping metrics: http://localhost:9090/targets
- Check Prometheus datasource connection in Grafana
- Verify metric names match in queries

### Slow Dashboard Performance
- Reduce time range
- Increase refresh interval
- Optimize PromQL queries
- Add recording rules in Prometheus

## Best Practices

1. **Use Recording Rules** - Pre-calculate complex queries in Prometheus
2. **Set Appropriate Refresh Intervals** - Balance freshness vs. load
3. **Use Template Variables** - Make dashboards reusable
4. **Add Annotations** - Mark deployments and incidents
5. **Export Regularly** - Keep dashboard JSON in version control

## Resources

- [Grafana Documentation](https://grafana.com/docs/)
- [Prometheus Query Language](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)

#