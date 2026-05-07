# Status and Error Handling Guide for IPQuorum Manager

## Overview
This guide explains how the IPQuorum Manager (central server) should interpret agent responses to determine instance health and get detailed error information.

---

## Status Interpretation Logic

### 1. Check Instance Status

```bash
GET /instances/{name}/status
```

**Response Scenarios:**

#### Scenario A: Service is Running (Healthy)
```json
{
  "name": "test-full",
  "status": "● ipquorum@test-full.service...\n     Active: active (running)..."
}
```
**Interpretation**: ✅ Instance is healthy and running

#### Scenario B: Service is Stopped (Intentional)
```json
{
  "error": "command failed: exit status 3\nstderr: "
}
```
**Interpretation**: ⏸️ Instance is stopped (exit code 3 = inactive)

#### Scenario C: Service Failed to Start
```json
{
  "error": "command failed: exit status 3\nstderr: "
}
```
**Interpretation**: ❌ Instance failed (need to check logs for details)

**Note**: Exit code 3 can mean either "stopped" or "failed" - you need to check logs to determine which.

---

## Getting Detailed Error Information

### Step 1: Check Status
```bash
curl -H "X-API-Key: $API_KEY" \
  http://localhost:9090/instances/test-full/status
```

### Step 2: If Error, Get Logs
```bash
curl -H "X-API-Key: $API_KEY" \
  "http://localhost:9090/instances/test-full/logs?lines=50"
```

**Logs will show the actual error**, for example:
- Connection refused (wrong IP/port)
- Authentication failed (wrong credentials)
- Java not found
- Port already in use
- Configuration error

---

## Recommended Manager Logic

### Python Example (for IPQuorum Manager)

```python
import requests
import json

class InstanceHealthChecker:
    def __init__(self, agent_url, api_key):
        self.agent_url = agent_url
        self.headers = {"X-API-Key": api_key}
    
    def get_instance_health(self, instance_name):
        """
        Get comprehensive instance health status
        Returns: dict with status, state, and error details
        """
        result = {
            "name": instance_name,
            "state": "unknown",
            "healthy": False,
            "status_output": None,
            "error_details": None,
            "logs": None
        }
        
        # Step 1: Get status
        try:
            status_resp = requests.get(
                f"{self.agent_url}/instances/{instance_name}/status",
                headers=self.headers,
                timeout=10
            )
            
            if status_resp.status_code == 200:
                data = status_resp.json()
                
                # Check if it's an error response
                if "error" in data:
                    # Exit code 3 = stopped or failed
                    if "exit status 3" in data["error"]:
                        result["state"] = "stopped_or_failed"
                        result["healthy"] = False
                        
                        # Get logs to determine if stopped or failed
                        logs = self._get_logs(instance_name, lines=50)
                        result["logs"] = logs
                        
                        # Parse logs for errors
                        if self._has_error_in_logs(logs):
                            result["state"] = "failed"
                            result["error_details"] = self._extract_error_from_logs(logs)
                        else:
                            result["state"] = "stopped"
                    else:
                        result["state"] = "error"
                        result["error_details"] = data["error"]
                
                # Check if running
                elif "status" in data:
                    status_text = data["status"]
                    result["status_output"] = status_text
                    
                    if "Active: active (running)" in status_text:
                        result["state"] = "running"
                        result["healthy"] = True
                        
                        # Check if actually connected
                        if "Connected to" in status_text:
                            result["state"] = "running_connected"
                        else:
                            result["state"] = "running_not_connected"
                            result["healthy"] = False
                            result["error_details"] = "Service running but not connected to storage"
                    
                    elif "Active: failed" in status_text:
                        result["state"] = "failed"
                        result["healthy"] = False
                        logs = self._get_logs(instance_name, lines=50)
                        result["logs"] = logs
                        result["error_details"] = self._extract_error_from_logs(logs)
            
            else:
                result["state"] = "api_error"
                result["error_details"] = f"HTTP {status_resp.status_code}"
        
        except requests.exceptions.RequestException as e:
            result["state"] = "unreachable"
            result["error_details"] = str(e)
        
        return result
    
    def _get_logs(self, instance_name, lines=50):
        """Get instance logs"""
        try:
            resp = requests.get(
                f"{self.agent_url}/instances/{instance_name}/logs",
                headers=self.headers,
                params={"lines": lines},
                timeout=10
            )
            if resp.status_code == 200:
                return resp.json().get("output", "")
        except:
            pass
        return None
    
    def _has_error_in_logs(self, logs):
        """Check if logs contain error indicators"""
        if not logs:
            return False
        
        error_keywords = [
            "failed",
            "error",
            "exception",
            "refused",
            "timeout",
            "cannot",
            "unable",
            "authentication failed",
            "connection refused"
        ]
        
        logs_lower = logs.lower()
        return any(keyword in logs_lower for keyword in error_keywords)
    
    def _extract_error_from_logs(self, logs):
        """Extract meaningful error message from logs"""
        if not logs:
            return "Unknown error - no logs available"
        
        # Look for common error patterns
        lines = logs.split('\n')
        error_lines = []
        
        for line in lines:
            line_lower = line.lower()
            if any(keyword in line_lower for keyword in 
                   ['error', 'failed', 'exception', 'refused', 'timeout']):
                error_lines.append(line.strip())
        
        if error_lines:
            # Return last few error lines (most recent)
            return '\n'.join(error_lines[-5:])
        
        # If no specific errors, return last few lines
        return '\n'.join(lines[-10:])


# Usage Example
checker = InstanceHealthChecker(
    agent_url="http://10.33.7.80:9090",
    api_key="ea17f6738a9db12f157265dc833154615f05c8c294269b4ced57d20684e6d4d7"
)

health = checker.get_instance_health("test-full")

print(f"Instance: {health['name']}")
print(f"State: {health['state']}")
print(f"Healthy: {health['healthy']}")

if health['error_details']:
    print(f"Error: {health['error_details']}")

if health['logs']:
    print(f"Recent logs:\n{health['logs'][:500]}...")
```

---

## State Machine

```
┌─────────────────────────────────────────────────────────────┐
│                     Instance States                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ✅ running_connected    - Healthy, connected to storage    │
│  ⚠️  running_not_connected - Running but not connected      │
│  ⏸️  stopped             - Intentionally stopped            │
│  ❌ failed               - Failed to start or crashed       │
│  ❓ unknown              - Cannot determine state           │
│  🔌 unreachable          - Agent not responding             │
│  ⚠️  api_error           - API returned error               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Common Error Scenarios

### 1. Authentication Failed
**Status**: Exit code 3  
**Logs contain**:
```
Authentication failed
Invalid credentials
401 Unauthorized
```
**Action**: Check username/password in config

### 2. Connection Refused
**Status**: Exit code 3  
**Logs contain**:
```
Connection refused
Cannot connect to 10.33.7.80:7443
Network unreachable
```
**Action**: Check API endpoint, firewall, network

### 3. Java Not Found
**Status**: Exit code 3  
**Logs contain**:
```
java: command not found
/usr/bin/java: No such file or directory
```
**Action**: Install Java on the host

### 4. Port Already in Use
**Status**: Exit code 3  
**Logs contain**:
```
Address already in use
Port 3260 is already in use
```
**Action**: Stop conflicting service or change port

### 5. Configuration Error
**Status**: Exit code 3  
**Logs contain**:
```
Invalid configuration
Missing required parameter
Cannot read config file
```
**Action**: Fix configuration file

---

## API Response Examples

### Healthy Instance
```json
{
  "name": "prod-cluster01",
  "state": "running_connected",
  "healthy": true,
  "status_output": "Active: active (running)...Connected to 10.33.7.90..."
}
```

### Failed Instance with Details
```json
{
  "name": "prod-cluster01",
  "state": "failed",
  "healthy": false,
  "error_details": "Authentication failed: Invalid credentials",
  "logs": "Apr 30 15:00:00 host ipquorum[1234]: ERROR: Authentication failed\nApr 30 15:00:00 host ipquorum[1234]: Invalid username or password..."
}
```

### Stopped Instance
```json
{
  "name": "prod-cluster01",
  "state": "stopped",
  "healthy": false,
  "logs": "Apr 30 14:55:00 host systemd[1]: Stopped IPQuorum service"
}
```

---

## Manager Dashboard Display

### Status Badge Colors

```
🟢 running_connected     → Green  "Running"
🟡 running_not_connected → Yellow "Running (Not Connected)"
⚪ stopped               → Gray   "Stopped"
🔴 failed                → Red    "Failed"
⚫ unknown               → Black  "Unknown"
🔌 unreachable           → Orange "Agent Unreachable"
```

### Error Display

When `state == "failed"`, show:
1. **Status Badge**: 🔴 Failed
2. **Error Message**: `error_details` (first 200 chars)
3. **Action Button**: "View Logs" → Shows full logs in modal
4. **Quick Actions**: 
   - "Restart Instance"
   - "View Configuration"
   - "Check Agent Logs"

---

## Polling Strategy

### Recommended Polling Intervals

```python
POLLING_INTERVALS = {
    "running_connected": 60,      # 1 minute
    "running_not_connected": 30,  # 30 seconds
    "stopped": 300,               # 5 minutes (no need to check often)
    "failed": 30,                 # 30 seconds (might recover)
    "unknown": 30,                # 30 seconds
    "unreachable": 60,            # 1 minute
}
```

### Smart Polling

```python
def get_next_poll_interval(current_state, consecutive_failures=0):
    """
    Adjust polling based on state and failure count
    """
    base_interval = POLLING_INTERVALS.get(current_state, 60)
    
    # Back off on consecutive failures
    if consecutive_failures > 0:
        backoff = min(consecutive_failures * 30, 300)  # Max 5 minutes
        return base_interval + backoff
    
    return base_interval
```

---

## Summary

**For IPQuorum Manager to detect errors:**

1. **Call status endpoint** → Get initial state
2. **If exit code 3** → Call logs endpoint
3. **Parse logs** → Extract error details
4. **Display to user** → Show state + error message
5. **Provide actions** → Restart, view config, check logs

**Key Points:**
- Status alone isn't enough - always check logs for errors
- Exit code 3 = stopped OR failed (logs tell you which)
- Logs contain the actual error messages
- Poll more frequently when instances are unhealthy
- Show actionable error messages to users
