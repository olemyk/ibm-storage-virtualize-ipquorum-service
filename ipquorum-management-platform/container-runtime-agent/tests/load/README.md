# Load Testing Guide

This directory contains load testing scripts for the Container Runtime Agent using k6.

## Prerequisites

Install k6:

```bash
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Docker
docker pull grafana/k6
```

## Running Load Tests

### Basic Load Test

```bash
# Start the agent first
cd ../..
go run main.go

# In another terminal, run the load test
cd tests/load
k6 run k6-load-test.js
```

### Custom Configuration

```bash
# Custom base URL
k6 run --env BASE_URL=http://localhost:8080 k6-load-test.js

# Custom credentials
k6 run --env USERNAME=admin --env PASSWORD=secret k6-load-test.js

# Custom duration and users
k6 run --stage 1m:50,2m:100,1m:0 k6-load-test.js
```

### With Docker

```bash
docker run --rm -i --network=host \
  -e BASE_URL=http://localhost:8080 \
  -e USERNAME=admin \
  -e PASSWORD=admin123 \
  grafana/k6 run - < k6-load-test.js
```

## Test Scenarios

The load test includes the following scenarios:

1. **Authentication** - Login and token retrieval
2. **List Containers** - Most common read operation
3. **Runtime Info** - System information retrieval
4. **Runtime Version** - Version information
5. **List Images** - Image listing
6. **Container Lifecycle** - Create, start, stop, remove (10% of iterations)

## Load Profile

Default load profile:
- **Ramp-up**: 30s to 10 users
- **Ramp-up**: 1m to 50 users
- **Ramp-up**: 2m to 100 users
- **Sustain**: 2m at 100 users
- **Ramp-down**: 1m to 50 users
- **Ramp-down**: 30s to 0 users

Total duration: ~7 minutes

## Performance Thresholds

- **95th percentile**: < 500ms
- **99th percentile**: < 1000ms
- **Error rate**: < 5%

## Metrics

The test tracks:
- `http_req_duration` - Request duration
- `http_req_failed` - Failed requests
- `errors` - Custom error rate
- `auth_duration` - Authentication time
- `container_list_duration` - Container listing time
- `container_create_duration` - Container creation time
- `requests_total` - Total request count

## Output

### Console Output

```
     ✓ login status is 200
     ✓ login has token
     ✓ list containers status is 200
     ✓ list containers returns array

     checks.........................: 100.00% ✓ 12000      ✗ 0
     data_received..................: 2.4 MB  340 kB/s
     data_sent......................: 1.2 MB  170 kB/s
     http_req_duration..............: avg=45ms    min=10ms med=40ms max=500ms p(90)=80ms p(95)=120ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 3000
     http_reqs......................: 3000    428/s
     iteration_duration.............: avg=7s      min=5s   med=7s   max=10s   p(90)=8s   p(95)=9s
     iterations.....................: 500     71/s
     vus............................: 100     min=0        max=100
     vus_max........................: 100     min=100      max=100
```

### JSON Output

```bash
k6 run --out json=results.json k6-load-test.js
```

### InfluxDB Output

```bash
k6 run --out influxdb=http://localhost:8086/k6 k6-load-test.js
```

### Grafana Dashboard

```bash
# Start InfluxDB and Grafana
docker-compose up -d influxdb grafana

# Run test with InfluxDB output
k6 run --out influxdb=http://localhost:8086/k6 k6-load-test.js

# View results in Grafana at http://localhost:3000
```

## Interpreting Results

### Good Performance
- p95 < 500ms
- p99 < 1000ms
- Error rate < 1%
- Throughput > 100 req/s

### Acceptable Performance
- p95 < 1000ms
- p99 < 2000ms
- Error rate < 5%
- Throughput > 50 req/s

### Poor Performance
- p95 > 1000ms
- p99 > 2000ms
- Error rate > 5%
- Throughput < 50 req/s

## Troubleshooting

### High Error Rate

Check:
- Agent logs for errors
- Database connection pool size
- Rate limiting configuration
- System resources (CPU, memory)

### High Latency

Check:
- Connection pool configuration
- Cache hit rate
- Database query performance
- Network latency

### Resource Exhaustion

Monitor:
- CPU usage
- Memory usage
- File descriptors
- Network connections

## Advanced Testing

### Spike Test

```bash
k6 run --stage 0s:0,10s:100,1m:100,10s:0 k6-load-test.js
```

### Stress Test

```bash
k6 run --stage 2m:100,5m:200,2m:300,5m:300,2m:200,2m:100,2m:0 k6-load-test.js
```

### Soak Test

```bash
k6 run --stage 5m:100,8h:100,5m:0 k6-load-test.js
```

## Cleanup

The test automatically cleans up test containers in the teardown phase. To manually clean up:

```bash
# Remove all test containers
docker ps -a --filter "label=test=k6-load-test" -q | xargs docker rm -f
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run Load Tests
  run: |
    k6 run --out json=results.json tests/load/k6-load-test.js
    
- name: Upload Results
  uses: actions/upload-artifact@v2
  with:
    name: k6-results
    path: results.json
```

### GitLab CI

```yaml
load_test:
  image: grafana/k6
  script:
    - k6 run --out json=results.json tests/load/k6-load-test.js
  artifacts:
    paths:
      - results.json
```

## Best Practices

1. **Baseline**: Establish baseline performance before changes
2. **Isolation**: Run tests in isolated environment
3. **Monitoring**: Monitor system resources during tests
4. **Repeatability**: Run tests multiple times for consistency
5. **Documentation**: Document test results and findings
6. **Thresholds**: Set realistic performance thresholds
7. **Cleanup**: Always clean up test data

## References

- [k6 Documentation](https://k6.io/docs/)
- [k6 Examples](https://k6.io/docs/examples/)
- [Performance Testing Best Practices](https://k6.io/docs/testing-guides/test-types/)