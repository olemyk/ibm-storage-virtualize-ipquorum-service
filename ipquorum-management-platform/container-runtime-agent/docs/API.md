# IPQuorum Container Runtime Agent API Documentation

## Table of Contents
- [Overview](#overview)
- [Authentication](#authentication)
- [Base URL](#base-url)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Endpoints](#endpoints)
  - [Authentication](#authentication-endpoints)
  - [Containers](#container-endpoints)
  - [Images](#image-endpoints)
  - [Runtime](#runtime-endpoints)
  - [Metrics](#metrics-endpoints)

---

## Overview

The IPQuorum Container Runtime Agent provides a REST API for managing containers through Docker or Podman runtimes. The API supports:

- JWT-based authentication
- Container lifecycle management (create, start, stop, restart, remove)
- Image management (list, pull)
- Runtime monitoring and health checks
- Prometheus metrics export

**API Version**: 1.0.0  
**Content-Type**: `application/json`  
**Authentication**: JWT Bearer Token

---

## Authentication

Most endpoints require JWT authentication. Include the token in the `Authorization` header:

```http
Authorization: Bearer <your-jwt-token>
```

### Obtaining a Token

Use the `/auth/login` endpoint to obtain a JWT token:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-01T12:00:00Z",
  "user": {
    "id": "123",
    "username": "admin",
    "roles": ["admin"]
  }
}
```

### Token Expiration

Tokens expire after 24 hours by default. Use the `/auth/refresh` endpoint to obtain a new token before expiration.

---

## Base URL

**Development**: `http://localhost:8080`  
**Production**: `https://agent.ipquorum.io`

---

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (success with no body) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

### Error Codes

| Code | Description |
|------|-------------|
| `BAD_REQUEST` | Invalid request parameters |
| `UNAUTHORIZED` | Authentication required or invalid |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource already exists |
| `RATE_LIMITED` | Too many requests |
| `INTERNAL_ERROR` | Server error |
| `RUNTIME_ERROR` | Container runtime error |

---

## Rate Limiting

Rate limits are applied per IP address:

- **Global**: 100 requests/minute
- **Per IP**: 50 requests/minute
- **Auth endpoints**: 10 requests/minute
- **Container operations**: 30 requests/minute

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

---

## Endpoints

### Authentication Endpoints

#### POST /auth/login

Authenticate user and receive JWT token.

**Request**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response** (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-01T12:00:00Z",
  "user": {
    "id": "123",
    "username": "admin",
    "roles": ["admin"]
  }
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

#### POST /auth/refresh

Refresh JWT token.

**Headers**: `Authorization: Bearer <token>`

**Response** (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-01T12:00:00Z",
  "user": {
    "id": "123",
    "username": "admin",
    "roles": ["admin"]
  }
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Authorization: Bearer $TOKEN"
```

---

#### POST /auth/logout

Logout and invalidate token.

**Headers**: `Authorization: Bearer <token>`

**Response** (200):
```json
{
  "message": "Logged out successfully"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

---

#### GET /auth/me

Get current user information.

**Headers**: `Authorization: Bearer <token>`

**Response** (200):
```json
{
  "id": "123",
  "username": "admin",
  "roles": ["admin"]
}
```

**Example**:
```bash
curl http://localhost:8080/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

---

### Container Endpoints

#### GET /containers

List all containers.

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `all` (boolean): Show all containers (default: false, shows only running)
- `filter` (string): Filter by name, status, etc.

**Response** (200):
```json
[
  {
    "id": "abc123",
    "name": "my-container",
    "image": "nginx:latest",
    "status": "running",
    "state": "running",
    "created": "2024-01-01T10:00:00Z",
    "started": "2024-01-01T10:00:05Z",
    "ports": [
      {
        "host_port": 8080,
        "container_port": 80,
        "protocol": "tcp"
      }
    ],
    "runtime": "docker"
  }
]
```

**Examples**:
```bash
# List running containers
curl http://localhost:8080/containers \
  -H "Authorization: Bearer $TOKEN"

# List all containers
curl "http://localhost:8080/containers?all=true" \
  -H "Authorization: Bearer $TOKEN"

# Filter by name
curl "http://localhost:8080/containers?filter=nginx" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### POST /containers

Create a new container.

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "name": "my-app",
  "image": "nginx:latest",
  "cmd": ["nginx", "-g", "daemon off;"],
  "env": [
    "ENV=production",
    "PORT=8080"
  ],
  "ports": [
    {
      "host_port": 8080,
      "container_port": 80,
      "protocol": "tcp"
    }
  ],
  "volumes": [
    {
      "source": "/host/path",
      "destination": "/container/path",
      "read_only": false
    }
  ],
  "labels": {
    "app": "my-app",
    "env": "production"
  },
  "restart_policy": "always",
  "resources": {
    "cpu_limit": "1.0",
    "memory_limit": "512m"
  }
}
```

**Response** (201):
```json
{
  "id": "abc123",
  "name": "my-app",
  "image": "nginx:latest",
  "status": "created",
  "state": "created",
  "created": "2024-01-01T10:00:00Z",
  "runtime": "docker"
}
```

**Examples**:
```bash
# Simple container
curl -X POST http://localhost:8080/containers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-nginx",
    "image": "nginx:latest"
  }'

# Container with environment and ports
curl -X POST http://localhost:8080/containers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app",
    "image": "node:18",
    "cmd": ["node", "server.js"],
    "env": ["NODE_ENV=production", "PORT=3000"],
    "ports": [{"host_port": 3000, "container_port": 3000, "protocol": "tcp"}]
  }'
```

---

#### GET /containers/{id}

Get detailed information about a container.

**Headers**: `Authorization: Bearer <token>`

**Path Parameters**:
- `id` (string): Container ID or name

**Response** (200):
```json
{
  "id": "abc123",
  "name": "my-container",
  "image": "nginx:latest",
  "status": "running",
  "state": "running",
  "created": "2024-01-01T10:00:00Z",
  "started": "2024-01-01T10:00:05Z",
  "exit_code": 0,
  "labels": {
    "app": "nginx"
  },
  "ports": [
    {
      "host_port": 8080,
      "container_port": 80,
      "protocol": "tcp"
    }
  ],
  "runtime": "docker"
}
```

**Example**:
```bash
curl http://localhost:8080/containers/my-container \
  -H "Authorization: Bearer $TOKEN"
```

---

#### DELETE /containers/{id}

Remove a container.

**Headers**: `Authorization: Bearer <token>`

**Path Parameters**:
- `id` (string): Container ID or name

**Query Parameters**:
- `force` (boolean): Force remove running container (default: false)

**Response** (204): No content

**Examples**:
```bash
# Remove stopped container
curl -X DELETE http://localhost:8080/containers/my-container \
  -H "Authorization: Bearer $TOKEN"

# Force remove running container
curl -X DELETE "http://localhost:8080/containers/my-container?force=true" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### POST /containers/{id}/start

Start a stopped container.

**Headers**: `Authorization: Bearer <token>`

**Path Parameters**:
- `id` (string): Container ID or name

**Response** (204): No content

**Example**:
```bash
curl -X POST http://localhost:8080/containers/my-container/start \
  -H "Authorization: Bearer $TOKEN"
```

---

#### POST /containers/{id}/stop

Stop a running container.

**Headers**: `Authorization: Bearer <token>`

**Path Parameters**:
- `id` (string): Container ID or name

**Query Parameters**:
- `timeout` (integer): Timeout in seconds before killing (default: 10)

**Response** (204): No content

**Example**:
```bash
# Stop with default timeout
curl -X POST http://localhost:8080/containers/my-container/stop \
  -H "Authorization: Bearer $TOKEN"

# Stop with custom timeout
curl -X POST "http://localhost:8080/containers/my-container/stop?timeout=30" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### POST /containers/{id}/restart

Restart a container.

**Headers**: `Authorization: Bearer <token>`

**Path Parameters**:
- `id` (string): Container ID or name

**Query Parameters**:
- `timeout` (integer): Timeout in seconds before killing (default: 10)

**Response** (204): No content

**Example**:
```bash
curl -X POST http://localhost:8080/containers/my-container/restart \
  -H "Authorization: Bearer $TOKEN"
```

---

#### GET /containers/{id}/logs

Get container logs.

**Headers**: `Authorization: Bearer <token>`

**Path Parameters**:
- `id` (string): Container ID or name

**Query Parameters**:
- `follow` (boolean): Follow log output (default: false)
- `tail` (integer): Number of lines from end (default: 100)
- `timestamps` (boolean): Show timestamps (default: false)

**Response** (200): Plain text logs

**Examples**:
```bash
# Get last 100 lines
curl http://localhost:8080/containers/my-container/logs \
  -H "Authorization: Bearer $TOKEN"

# Get last 50 lines with timestamps
curl "http://localhost:8080/containers/my-container/logs?tail=50&timestamps=true" \
  -H "Authorization: Bearer $TOKEN"

# Follow logs (streaming)
curl "http://localhost:8080/containers/my-container/logs?follow=true" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### GET /containers/{id}/stats

Get container resource usage statistics.

**Headers**: `Authorization: Bearer <token>`

**Path Parameters**:
- `id` (string): Container ID or name

**Response** (200):
```json
{
  "cpu_percent": 15.5,
  "memory_usage": 134217728,
  "memory_limit": 536870912,
  "memory_percent": 25.0,
  "network_rx": 1024000,
  "network_tx": 512000,
  "block_read": 2048000,
  "block_write": 1024000,
  "pids": 10
}
```

**Example**:
```bash
curl http://localhost:8080/containers/my-container/stats \
  -H "Authorization: Bearer $TOKEN"
```

---

### Image Endpoints

#### GET /images

List all images.

**Headers**: `Authorization: Bearer <token>`

**Response** (200):
```json
[
  {
    "id": "sha256:abc123...",
    "tags": ["nginx:latest", "nginx:1.21"],
    "size": 142000000,
    "created": "2024-01-01T10:00:00Z"
  }
]
```

**Example**:
```bash
curl http://localhost:8080/images \
  -H "Authorization: Bearer $TOKEN"
```

---

#### POST /images/pull

Pull an image from registry.

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "image": "nginx:latest"
}
```

**Response** (200):
```json
{
  "message": "Image pulled successfully"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/images/pull \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"image":"nginx:latest"}'
```

---

### Runtime Endpoints

#### GET /runtime/info

Get runtime system information.

**Headers**: `Authorization: Bearer <token>`

**Response** (200):
```json
{
  "name": "Docker",
  "type": "docker",
  "version": "24.0.0",
  "api_version": "1.43",
  "os": "Linux",
  "arch": "x86_64",
  "kernel_version": "5.15.0",
  "containers_running": 5,
  "containers_paused": 0,
  "containers_stopped": 3,
  "images": 10,
  "driver": "overlay2",
  "memory_total": 16777216000,
  "cpus": 8
}
```

**Example**:
```bash
curl http://localhost:8080/runtime/info \
  -H "Authorization: Bearer $TOKEN"
```

---

#### GET /runtime/version

Get runtime version information.

**Headers**: `Authorization: Bearer <token>`

**Response** (200):
```json
{
  "version": "24.0.0",
  "api_version": "1.43",
  "min_api_version": "1.12",
  "go_version": "go1.20.0",
  "git_commit": "abc123",
  "build_time": "2024-01-01T10:00:00Z",
  "os": "linux",
  "arch": "amd64"
}
```

**Example**:
```bash
curl http://localhost:8080/runtime/version \
  -H "Authorization: Bearer $TOKEN"
```

---

#### GET /health

Health check endpoint (no authentication required).

**Response** (200):
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T10:00:00Z",
  "runtime": "docker",
  "uptime": 3600
}
```

**Example**:
```bash
curl http://localhost:8080/health
```

---

### Metrics Endpoints

#### GET /metrics

Get Prometheus metrics (no authentication required).

**Response** (200): Prometheus text format

**Example**:
```bash
curl http://localhost:8080/metrics
```

**Sample Output**:
```
# HELP ipquorum_agent_runtime_available Runtime availability
# TYPE ipquorum_agent_runtime_available gauge
ipquorum_agent_runtime_available{runtime="docker"} 1

# HELP ipquorum_agent_containers_running Number of running containers
# TYPE ipquorum_agent_containers_running gauge
ipquorum_agent_containers_running{runtime="docker"} 5

# HELP ipquorum_agent_api_requests_total Total API requests
# TYPE ipquorum_agent_api_requests_total counter
ipquorum_agent_api_requests_total{endpoint="/containers",method="GET",status="200"} 42
```

---

## Complete Examples

### Complete Container Lifecycle

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# 2. Create container
CONTAINER_ID=$(curl -s -X POST http://localhost:8080/containers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-nginx",
    "image": "nginx:latest",
    "ports": [{"host_port": 8080, "container_port": 80, "protocol": "tcp"}]
  }' | jq -r '.id')

# 3. Start container
curl -X POST http://localhost:8080/containers/$CONTAINER_ID/start \
  -H "Authorization: Bearer $TOKEN"

# 4. Get container info
curl http://localhost:8080/containers/$CONTAINER_ID \
  -H "Authorization: Bearer $TOKEN"

# 5. Get logs
curl http://localhost:8080/containers/$CONTAINER_ID/logs \
  -H "Authorization: Bearer $TOKEN"

# 6. Get stats
curl http://localhost:8080/containers/$CONTAINER_ID/stats \
  -H "Authorization: Bearer $TOKEN"

# 7. Stop container
curl -X POST http://localhost:8080/containers/$CONTAINER_ID/stop \
  -H "Authorization: Bearer $TOKEN"

# 8. Remove container
curl -X DELETE http://localhost:8080/containers/$CONTAINER_ID \
  -H "Authorization: Bearer $TOKEN"
```

---

## SDK Examples

### Go
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func main() {
    // Login
    loginReq := map[string]string{
        "username": "admin",
        "password": "admin123",
    }
    body, _ := json.Marshal(loginReq)
    resp, _ := http.Post("http://localhost:8080/auth/login", 
        "application/json", bytes.NewBuffer(body))
    
    var loginResp struct {
        Token string `json:"token"`
    }
    json.NewDecoder(resp.Body).Decode(&loginResp)
    
    // Create container
    req, _ := http.NewRequest("POST", "http://localhost:8080/containers", nil)
    req.Header.Set("Authorization", "Bearer "+loginResp.Token)
    // ... rest of implementation
}
```

### Python
```python
import requests

# Login
response = requests.post('http://localhost:8080/auth/login', json={
    'username': 'admin',
    'password': 'admin123'
})
token = response.json()['token']

# Create container
headers = {'Authorization': f'Bearer {token}'}
response = requests.post('http://localhost:8080/containers', 
    headers=headers,
    json={
        'name': 'my-nginx',
        'image': 'nginx:latest'
    })
container = response.json()
```

### JavaScript/Node.js
```javascript
const axios = require('axios');

async function main() {
    // Login
    const loginResp = await axios.post('http://localhost:8080/auth/login', {
        username: 'admin',
        password: 'admin123'
    });
    const token = loginResp.data.token;
    
    // Create container
    const containerResp = await axios.post('http://localhost:8080/containers', {
        name: 'my-nginx',
        image: 'nginx:latest'
    }, {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    console.log(containerResp.data);
}
```

---

## Troubleshooting

### Common Issues

**401 Unauthorized**
- Check token is valid and not expired
- Ensure Authorization header is correctly formatted
- Try refreshing the token

**404 Not Found**
- Verify container ID/name is correct
- Check container exists with GET /containers

**500 Internal Server Error**
- Check Docker/Podman daemon is running
- Verify agent has permissions to access runtime socket
- Check agent logs for details

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=debug
./agent
```

---

## Support

- **Documentation**: https://docs.ipquorum.io
- **Issues**: https://github.com/ipquorum/agent/issues
- **Email**: support@ipquorum.io