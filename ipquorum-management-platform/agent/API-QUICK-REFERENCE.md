# Agent API Quick Reference

## Base URL
```
http://localhost:9090
```

## Authentication
All requests (except `/health`) require API key in header:
```
X-API-Key: your-api-key-here
```

---

## Endpoints

### Health Check
```bash
curl http://localhost:9090/health
```

### List Instances
```bash
curl -H "X-API-Key: YOUR_KEY" \
  http://localhost:9090/instances
```

### Create Instance (Minimal)
```bash
curl -X POST http://localhost:9090/instances \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cluster01",
    "api_endpoint": "10.33.7.80",
    "username": "monitor",
    "password": "secret123"
  }'
```

### Create Instance (Full)
```bash
curl -X POST http://localhost:9090/instances \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cluster01",
    "api_endpoint": "10.33.7.80",
    "username": "admin",
    "password": "secret123",
    "storage_system": "svc_cluster01",
    "storage_description": "IBM FlashSystem 9200",
    "storage_location": "Datacenter A",
    "mkquorumapp_enabled": true,
    "partnersystem": "svc_cluster02",
    "ip6": false,
    "partnerip6": false,
    "nometadata": false
  }'
```

### Get Instance Status
```bash
curl -H "X-API-Key: YOUR_KEY" \
  http://localhost:9090/instances/INSTANCE_NAME/status
```

### Get Instance Logs
```bash
# Default: 100 lines
curl -H "X-API-Key: YOUR_KEY" \
  http://localhost:9090/instances/INSTANCE_NAME/logs

# Custom line count
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:9090/instances/INSTANCE_NAME/logs?lines=50"
```

### Start Instance ⚠️ **USE POST**
```bash
curl -X POST http://localhost:9090/instances/INSTANCE_NAME/start \
  -H "X-API-Key: YOUR_KEY"
```

### Stop Instance ⚠️ **USE POST**
```bash
curl -X POST http://localhost:9090/instances/INSTANCE_NAME/stop \
  -H "X-API-Key: YOUR_KEY"
```

### Restart Instance ⚠️ **USE POST**
```bash
curl -X POST http://localhost:9090/instances/INSTANCE_NAME/restart \
  -H "X-API-Key: YOUR_KEY"
```

### Delete Instance
```bash
curl -X DELETE http://localhost:9090/instances/INSTANCE_NAME \
  -H "X-API-Key: YOUR_KEY"
```

---

## Common Mistakes

### ❌ WRONG - Using GET for start/stop/restart
```bash
# This will return 404
curl -H "X-API-Key: YOUR_KEY" \
  http://localhost:9090/instances/test-full/stop
```

### ✅ CORRECT - Using POST for start/stop/restart
```bash
# This works
curl -X POST http://localhost:9090/instances/test-full/stop \
  -H "X-API-Key: YOUR_KEY"
```

---

## HTTP Methods Summary

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check (no auth) |
| `/instances` | GET | List all instances |
| `/instances` | POST | Create new instance |
| `/instances/:name/status` | GET | Get instance status |
| `/instances/:name/logs` | GET | Get instance logs |
| `/instances/:name/start` | **POST** | Start instance |
| `/instances/:name/stop` | **POST** | Stop instance |
| `/instances/:name/restart` | **POST** | Restart instance |
| `/instances/:name` | DELETE | Delete instance |

---

## Examples with Your Setup

Replace `YOUR_KEY` with your actual API key:
```bash
API_KEY="ea17f6738a9db12f157265dc833154615f05c8c294269b4ced57d20684e6d4d7"
```

### Stop Instance
```bash
curl -X POST http://localhost:9090/instances/test-full/stop \
  -H "X-API-Key: $API_KEY"
```

### Start Instance
```bash
curl -X POST http://localhost:9090/instances/test-full/start \
  -H "X-API-Key: $API_KEY"
```

### Restart Instance
```bash
curl -X POST http://localhost:9090/instances/test-full/restart \
  -H "X-API-Key: $API_KEY"
```

### Get Status
```bash
curl -H "X-API-Key: $API_KEY" \
  http://localhost:9090/instances/test-full/status
```

### Get Logs
```bash
curl -H "X-API-Key: $API_KEY" \
  http://localhost:9090/instances/test-full/logs
```

---

## Response Examples

### Success Response
```json
{
  "message": "Instance stopped successfully"
}
```

### Error Response
```json
{
  "error": "Instance not found"
}
```

---

## Testing Script

Save this as `test-instance-operations.sh`:

```bash
#!/bin/bash
API_KEY="your-api-key"
INSTANCE="test-full"
BASE_URL="http://localhost:9090"

echo "=== Testing Instance Operations ==="

echo "1. Get Status"
curl -H "X-API-Key: $API_KEY" \
  "$BASE_URL/instances/$INSTANCE/status"
echo -e "\n"

echo "2. Stop Instance"
curl -X POST "$BASE_URL/instances/$INSTANCE/stop" \
  -H "X-API-Key: $API_KEY"
echo -e "\n"

sleep 2

echo "3. Get Status (should be stopped)"
curl -H "X-API-Key: $API_KEY" \
  "$BASE_URL/instances/$INSTANCE/status"
echo -e "\n"

echo "4. Start Instance"
curl -X POST "$BASE_URL/instances/$INSTANCE/start" \
  -H "X-API-Key: $API_KEY"
echo -e "\n"

sleep 2

echo "5. Get Status (should be running)"
curl -H "X-API-Key: $API_KEY" \
  "$BASE_URL/instances/$INSTANCE/status"
echo -e "\n"

echo "6. Restart Instance"
curl -X POST "$BASE_URL/instances/$INSTANCE/restart" \
  -H "X-API-Key: $API_KEY"
echo -e "\n"

echo "7. Get Logs (last 20 lines)"
curl -H "X-API-Key: $API_KEY" \
  "$BASE_URL/instances/$INSTANCE/logs?lines=20"
echo -e "\n"
```

---

## Troubleshooting

### 404 Not Found
- **Cause**: Using GET instead of POST for start/stop/restart
- **Fix**: Add `-X POST` to your curl command

### 401 Unauthorized
- **Cause**: Missing or invalid API key
- **Fix**: Check your API key in `/etc/ipquorum/agent.yaml`

### 500 Internal Server Error
- **Cause**: Script execution failed
- **Fix**: Check agent logs: `journalctl -u ipquorum-agent -f`

---

## Quick Copy-Paste Commands

```bash
# Set your API key
export API_KEY="ea17f6738a9db12f157265dc833154615f05c8c294269b4ced57d20684e6d4d7"

# Stop
curl -X POST http://localhost:9090/instances/test-full/stop -H "X-API-Key: $API_KEY"

# Start
curl -X POST http://localhost:9090/instances/test-full/start -H "X-API-Key: $API_KEY"

# Restart
curl -X POST http://localhost:9090/instances/test-full/restart -H "X-API-Key: $API_KEY"

# Status
curl -H "X-API-Key: $API_KEY" http://localhost:9090/instances/test-full/status

# Logs
curl -H "X-API-Key: $API_KEY" http://localhost:9090/instances/test-full/logs