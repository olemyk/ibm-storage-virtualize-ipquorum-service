# IP Quorum Management Platform - API Specification

## Overview

**Version:** 1.0.0  
**Base URL:** `http://localhost:8080`  
**API Prefix:** `/api/v1`

This document describes the REST API for the IP Quorum Management Platform.

## Authentication

All protected endpoints require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

### Obtaining a Token

**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "username": "admin",
  "password": "changeme"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "uuid",
    "username": "admin",
    "role": "admin"
  }
}
```

## Roles

- **admin**: Full access to all operations
- **operator**: Can manage instances but not users/config
- **viewer**: Read-only access

## Endpoints

### Health & Monitoring

#### GET /health

Check server health status (no authentication required).

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-04-26T19:00:00Z",
  "components": {
    "database": "healthy",
    "instances": {
      "total": 5,
      "healthy": 3,
      "unhealthy": 1,
      "degraded": 1
    }
  }
}
```

#### GET /metrics

Prometheus metrics endpoint (no authentication required).

**Response:** Prometheus text format

---

### Authentication

#### POST /api/v1/auth/login

Authenticate user and obtain JWT token.

**Request:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:** `200 OK`
```json
{
  "token": "string",
  "refresh_token": "string",
  "expires_in": 86400,
  "user": {
    "id": "uuid",
    "username": "string",
    "role": "string"
  }
}
```

**Errors:**
- `401 Unauthorized`: Invalid credentials

#### POST /api/v1/auth/logout

Logout current user (invalidate token).

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

#### POST /api/v1/auth/refresh

Refresh access token using refresh token.

**Request:**
```json
{
  "refresh_token": "string"
}
```

**Response:** `200 OK`
```json
{
  "token": "string",
  "expires_in": 86400
}
```

#### GET /api/v1/auth/me

Get current user information.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "username": "string",
  "role": "string",
  "created_at": "2026-04-26T19:00:00Z",
  "last_login": "2026-04-26T19:00:00Z"
}
```

---

### Instance Management

#### GET /api/v1/instances

List all IP Quorum instances.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `server_id` (optional): Filter by server ID
- `status` (optional): Filter by status (running, stopped, failed)
- `health` (optional): Filter by health (healthy, unhealthy, degraded)

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "instance-1",
    "server_id": "local",
    "api_endpoint": "10.0.0.100",
    "username": "superuser",
    "storage_system": "SVC Cluster 1",
    "description": "Production instance",
    "location": "Datacenter A",
    "status": "running",
    "health": "healthy",
    "enable_download": true,
    "enable_mkquorumapp": true,
    "partnersystem": "remote-cluster",
    "created_at": "2026-04-26T19:00:00Z",
    "updated_at": "2026-04-26T19:00:00Z"
  }
]
```

#### POST /api/v1/instances

Create a new IP Quorum instance.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin` or `operator`

**Request:**
```json
{
  "name": "instance-1",
  "server_id": "local",
  "api_endpoint": "10.0.0.100",
  "username": "superuser",
  "password": "password",
  "storage_system": "SVC Cluster 1",
  "description": "Production instance",
  "location": "Datacenter A",
  "enable_download": true,
  "enable_mkquorumapp": true,
  "partnersystem": "remote-cluster",
  "ipquorum_name": "ipq-prod-01",
  "ip6": false,
  "partnerip6": false,
  "nometadata": false
}
```

**Field Descriptions:**
- `name` (required): Instance identifier
- `api_endpoint` (required): Storage system IP or hostname
- `username` (required): Storage system username
- `password` (required): Storage system password
- `enable_download` (optional, default: true): Download ip_quorum.jar
- `enable_mkquorumapp` (optional, default: false): Create quorum application
- `partnersystem` (required if mkquorumapp enabled): Remote system name for PBHA
- `ipquorum_name` (optional): Custom IPQuorum application name (auto-generated if not provided)
- `ip6` (optional, default: false): Enable IPv6 for local system
- `partnerip6` (optional, default: false): Enable IPv6 for partner system
- `nometadata` (optional, default: false): Disable metadata collection

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "name": "instance-1",
  "server_id": "local",
  "api_endpoint": "10.0.0.100",
  "username": "superuser",
  "storage_system": "SVC Cluster 1",
  "description": "Production instance",
  "location": "Datacenter A",
  "status": "stopped",
  "health": "unknown",
  "enable_download": true,
  "enable_mkquorumapp": true,
  "partnersystem": "remote-cluster",
  "ipquorum_name": "ipq-prod-01",
  "ip6": false,
  "partnerip6": false,
  "nometadata": false,
  "created_at": "2026-04-26T19:00:00Z",
  "updated_at": "2026-04-26T19:00:00Z"
}
```

**Errors:**
- `400 Bad Request`: Invalid input
- `409 Conflict`: Instance name already exists

#### GET /api/v1/instances/:id

Get details of a specific instance.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "name": "instance-1",
  "server_id": "local",
  "api_endpoint": "10.0.0.100",
  "username": "superuser",
  "storage_system": "SVC Cluster 1",
  "description": "Production instance",
  "location": "Datacenter A",
  "status": "running",
  "health": "healthy",
  "enable_download": true,
  "enable_mkquorumapp": true,
  "partnersystem": "remote-cluster",
  "created_at": "2026-04-26T19:00:00Z",
  "updated_at": "2026-04-26T19:00:00Z"
}
```

**Errors:**
- `404 Not Found`: Instance not found

#### PUT /api/v1/instances/:id

Update an existing instance.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin` or `operator`

**Request:**
```json
{
  "description": "Updated description",
  "location": "Datacenter B",
  "storage_system": "SVC Cluster 2"
}
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "name": "instance-1",
  "description": "Updated description",
  "location": "Datacenter B",
  "storage_system": "SVC Cluster 2",
  ...
}
```

**Errors:**
- `404 Not Found`: Instance not found
- `400 Bad Request`: Invalid input

#### DELETE /api/v1/instances/:id

Delete an instance.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin` or `operator`

**Query Parameters:**
- `force` (optional): Force delete even if running (default: false)

**Response:** `200 OK`
```json
{
  "message": "Instance deleted successfully"
}
```

**Errors:**
- `404 Not Found`: Instance not found
- `409 Conflict`: Instance is running (use force=true)

---

### Instance Operations

#### POST /api/v1/instances/:id/start

Start an IP Quorum instance.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin` or `operator`

**Response:** `200 OK`
```json
{
  "message": "Instance started successfully"
}
```

**Errors:**
- `404 Not Found`: Instance not found
- `409 Conflict`: Instance already running

#### POST /api/v1/instances/:id/stop

Stop an IP Quorum instance.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin` or `operator`

**Response:** `200 OK`
```json
{
  "message": "Instance stopped successfully"
}
```

**Errors:**
- `404 Not Found`: Instance not found
- `409 Conflict`: Instance already stopped

#### POST /api/v1/instances/:id/restart

Restart an IP Quorum instance.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin` or `operator`

**Response:** `200 OK`
```json
{
  "message": "Instance restarted successfully"
}
```

**Errors:**
- `404 Not Found`: Instance not found

#### GET /api/v1/instances/:id/status

Get current status of an instance.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "name": "instance-1",
  "status": "running",
  "health": "healthy",
  "updated": "2026-04-26T19:00:00Z"
}
```

**Errors:**
- `404 Not Found`: Instance not found

#### GET /api/v1/instances/:id/logs

Get logs for an instance.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `lines` (optional): Number of log lines to return (default: 100)
- `follow` (optional): Stream logs (default: false)

**Response:** `200 OK`
```json
{
  "logs": "string (log content)",
  "lines": 100
}
```

**Errors:**
- `404 Not Found`: Instance not found

---

### Health Monitoring

#### GET /api/v1/health/instances

Get health status of all instances.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "total": 5,
  "healthy": 3,
  "unhealthy": 1,
  "degraded": 1,
  "instances": [
    {
      "id": "uuid",
      "name": "instance-1",
      "status": "running",
      "health": "healthy",
      "last_check": "2026-04-26T19:00:00Z"
    }
  ]
}
```

#### GET /api/v1/health/instances/:id

Get detailed health information for a specific instance.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "instance": {
    "id": "uuid",
    "name": "instance-1",
    "status": "running",
    "health": "healthy",
    "updated": "2026-04-26T19:00:00Z"
  },
  "history": [
    {
      "timestamp": "2026-04-26T19:00:00Z",
      "status": "running",
      "health": "healthy"
    }
  ]
}
```

**Errors:**
- `404 Not Found`: Instance not found

---

### Metrics (Protected)

#### GET /api/v1/metrics/instances/:id

Get metrics for a specific instance.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `from` (optional): Start time (ISO 8601)
- `to` (optional): End time (ISO 8601)

**Response:** `200 OK`
```json
{
  "instance_id": "uuid",
  "metrics": {
    "uptime_seconds": 86400,
    "health_checks_total": 2880,
    "health_checks_failed": 5,
    "operations_total": 150,
    "operations_failed": 2
  },
  "period": {
    "from": "2026-04-25T19:00:00Z",
    "to": "2026-04-26T19:00:00Z"
  }
}
```

**Errors:**
- `404 Not Found`: Instance not found

---

### Server Management

#### GET /api/v1/servers

List all registered servers.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin`

**Response:** `200 OK`
```json
[
  {
    "id": "local",
    "name": "Local Server",
    "hostname": "localhost",
    "status": "online",
    "instances_count": 5,
    "created_at": "2026-04-26T19:00:00Z"
  }
]
```

#### POST /api/v1/servers

Register a new server.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin`

**Request:**
```json
{
  "name": "Remote Server 1",
  "hostname": "remote.example.com",
  "port": 8080,
  "api_key": "string"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "name": "Remote Server 1",
  "hostname": "remote.example.com",
  "port": 8080,
  "status": "online",
  "created_at": "2026-04-26T19:00:00Z"
}
```

#### GET /api/v1/servers/:id

Get server details.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "name": "Remote Server 1",
  "hostname": "remote.example.com",
  "port": 8080,
  "status": "online",
  "instances_count": 3,
  "created_at": "2026-04-26T19:00:00Z",
  "last_seen": "2026-04-26T19:00:00Z"
}
```

#### DELETE /api/v1/servers/:id

Unregister a server.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin`

**Response:** `200 OK`
```json
{
  "message": "Server unregistered successfully"
}
```

#### GET /api/v1/servers/:id/instances

Get all instances on a specific server.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin`

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "instance-1",
    "status": "running",
    "health": "healthy"
  }
]
```

---

### Configuration

#### GET /api/v1/config

Get current configuration.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin`

**Response:** `200 OK`
```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "log_level": "info",
    "health_check_interval": 30
  },
  "auth": {
    "token_expiry": 86400,
    "refresh_expiry": 604800
  },
  "agent": {
    "scripts_dir": "/scripts"
  }
}
```

#### PUT /api/v1/config

Update configuration.

**Headers:** `Authorization: Bearer <token>`  
**Required Role:** `admin`

**Request:**
```json
{
  "server": {
    "log_level": "debug",
    "health_check_interval": 60
  }
}
```

**Response:** `200 OK`
```json
{
  "message": "Configuration updated successfully",
  "restart_required": false
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Additional details (optional)"
}
```

### Common HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., duplicate name)
- `500 Internal Server Error`: Server error

## Rate Limiting

API requests are rate-limited to prevent abuse:

- **Authenticated requests**: 1000 requests per hour per user
- **Unauthenticated requests**: 100 requests per hour per IP

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1619712000
```

## Pagination

List endpoints support pagination:

**Query Parameters:**
- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 20, max: 100)

**Response Headers:**
```
X-Total-Count: 150
X-Page: 1
X-Per-Page: 20
X-Total-Pages: 8
```

## Webhooks

Configure webhooks to receive notifications about instance events:

**Events:**
- `instance.created`
- `instance.updated`
- `instance.deleted`
- `instance.started`
- `instance.stopped`
- `instance.health_changed`

**Webhook Payload:**
```json
{
  "event": "instance.health_changed",
  "timestamp": "2026-04-26T19:00:00Z",
  "data": {
    "instance_id": "uuid",
    "instance_name": "instance-1",
    "old_health": "healthy",
    "new_health": "unhealthy"
  }
}
```

## SDK Examples

### cURL

```bash
# Login
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}' \
  | jq -r '.token')

# List instances
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/instances

# Create instance
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-instance",
    "api_endpoint": "10.0.0.100",
    "username": "superuser",
    "password": "password",
    "partnersystem": "remote-cluster"
  }'
```

### Python

```python
import requests

# Login
response = requests.post('http://localhost:8080/api/v1/auth/login', json={
    'username': 'admin',
    'password': 'changeme'
})
token = response.json()['token']

# List instances
headers = {'Authorization': f'Bearer {token}'}
response = requests.get('http://localhost:8080/api/v1/instances', headers=headers)
instances = response.json()
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
    // Login
    loginData := map[string]string{"username": "admin", "password": "changeme"}
    body, _ := json.Marshal(loginData)
    resp, _ := http.Post("http://localhost:8080/api/v1/auth/login", 
        "application/json", bytes.NewBuffer(body))
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    token := result["token"].(string)
    
    // List instances
    req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/instances", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    client := &http.Client{}
    resp, _ = client.Do(req)
}
```

## Changelog

### Version 1.0.0 (2026-04-26)
- Initial API release
- Authentication endpoints
- Instance management
- Health monitoring
- Metrics collection
- Server management
- Configuration management