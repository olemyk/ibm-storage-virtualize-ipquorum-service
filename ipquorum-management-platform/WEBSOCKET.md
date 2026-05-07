# IPQuorum Management Platform - WebSocket Real-Time Updates

This document describes the WebSocket implementation for real-time updates in the IPQuorum Management Platform.

## Overview

WebSocket provides bidirectional, real-time communication between the server and web dashboard, enabling:
- Live instance status updates
- Real-time metrics streaming
- Instant alert notifications
- System health monitoring

## Architecture

```
┌─────────────────┐
│  Web Dashboard  │
│   (React/TS)    │
└────────┬────────┘
         │ WebSocket
         │ ws://localhost:8080/ws
         │
         ▼
┌─────────────────┐
│   API Server    │
│   (Go/Gin)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  WebSocket Hub  │
│  (Gorilla WS)   │
└─────────────────┘
```

## Backend Implementation

### Dependencies

Add to `go.mod`:
```bash
cd server
go get github.com/gorilla/websocket
```

### WebSocket Hub

The hub manages all WebSocket connections and broadcasts messages.

**File:** `server/pkg/websocket/hub.go`

Key features:
- Client registration/unregistration
- Message broadcasting to all clients
- User-specific message routing
- Thread-safe operations

### Message Types

```go
type MessageType string

const (
    MessageTypeInstanceUpdate MessageType = "instance_update"
    MessageTypeMetricsUpdate  MessageType = "metrics_update"
    MessageTypeAlert          MessageType = "alert"
    MessageTypeHeartbeat      MessageType = "heartbeat"
)
```

### Message Format

```json
{
  "type": "instance_update",
  "timestamp": "2024-01-20T10:30:00Z",
  "data": {
    "instance_id": 1,
    "name": "ipquorum-01",
    "status": "running"
  }
}
```

### API Endpoint

**Endpoint:** `GET /ws`

**Authentication:** JWT token via query parameter or header

**Example:**
```
ws://localhost:8080/ws?token=<jwt_token>
```

### Server Integration

```go
// In main.go or server setup
hub := websocket.NewHub()
go hub.Run()

// Add WebSocket endpoint
router.GET("/ws", func(c *gin.Context) {
    websocket.ServeWs(hub, c.Writer, c.Request)
})

// Broadcast updates
hub.Broadcast(websocket.MessageTypeInstanceUpdate, instanceData)
```

## Frontend Implementation

### React WebSocket Hook

**File:** `web/src/hooks/useWebSocket.ts`

```typescript
import { useEffect, useRef, useState, useCallback } from 'react';

interface WebSocketMessage {
  type: string;
  timestamp: string;
  data: any;
}

interface UseWebSocketOptions {
  url: string;
  onMessage?: (message: WebSocketMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (error: Event) => void;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export const useWebSocket = (options: UseWebSocketOptions) => {
  const {
    url,
    onMessage,
    onOpen,
    onClose,
    onError,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options;

  const [isConnected, setIsConnected] = useState(false);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

  const connect = useCallback(() => {
    try {
      const token = localStorage.getItem('auth_token');
      const wsUrl = `${url}?token=${token}`;
      
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        setReconnectAttempts(0);
        onOpen?.();
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          onMessage?.(message);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        wsRef.current = null;
        onClose?.();

        // Attempt reconnection
        if (reconnectAttempts < maxReconnectAttempts) {
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log(`Reconnecting... (${reconnectAttempts + 1}/${maxReconnectAttempts})`);
            setReconnectAttempts(prev => prev + 1);
            connect();
          }, reconnectInterval);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        onError?.(error);
      };
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
    }
  }, [url, onMessage, onOpen, onClose, onError, reconnectAttempts, maxReconnectAttempts, reconnectInterval]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const send = useCallback((data: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
    } else {
      console.warn('WebSocket is not connected');
    }
  }, []);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  return {
    isConnected,
    send,
    disconnect,
    reconnect: connect,
  };
};
```

### Usage in Components

```typescript
// In Dashboard.tsx
import { useWebSocket } from '../hooks/useWebSocket';

const Dashboard: React.FC = () => {
  const [instances, setInstances] = useState([]);
  const [metrics, setMetrics] = useState({});

  const { isConnected } = useWebSocket({
    url: 'ws://localhost:8080/ws',
    onMessage: (message) => {
      switch (message.type) {
        case 'instance_update':
          // Update instance data
          setInstances(prev => updateInstance(prev, message.data));
          break;
        case 'metrics_update':
          // Update metrics
          setMetrics(message.data);
          break;
        case 'alert':
          // Show alert notification
          showToast(message.data.message, 'warning');
          break;
      }
    },
    onOpen: () => {
      console.log('Connected to real-time updates');
    },
    onClose: () => {
      console.log('Disconnected from real-time updates');
    },
  });

  return (
    <div>
      <div className="status-indicator">
        {isConnected ? '🟢 Live' : '🔴 Offline'}
      </div>
      {/* Dashboard content */}
    </div>
  );
};
```

## Message Types & Payloads

### Instance Update

```json
{
  "type": "instance_update",
  "timestamp": "2024-01-20T10:30:00Z",
  "data": {
    "instance_id": 1,
    "name": "ipquorum-01",
    "status": "running",
    "host": "10.0.0.1",
    "port": 3993,
    "last_health_check": "2024-01-20T10:29:55Z"
  }
}
```

### Metrics Update

```json
{
  "type": "metrics_update",
  "timestamp": "2024-01-20T10:30:00Z",
  "data": {
    "total_instances": 5,
    "active_instances": 3,
    "api_requests_per_second": 45.2,
    "avg_response_time_ms": 125,
    "error_rate": 0.02
  }
}
```

### Alert Notification

```json
{
  "type": "alert",
  "timestamp": "2024-01-20T10:30:00Z",
  "data": {
    "severity": "warning",
    "title": "High API Latency",
    "message": "API response time exceeded 1 second",
    "instance_id": 2
  }
}
```

### Heartbeat

```json
{
  "type": "heartbeat",
  "timestamp": "2024-01-20T10:30:00Z",
  "data": {
    "server_time": "2024-01-20T10:30:00Z",
    "connected_clients": 5
  }
}
```

## Security

### Authentication

- JWT token required for WebSocket connection
- Token passed via query parameter or Sec-WebSocket-Protocol header
- Token validated on connection establishment

### Authorization

- User-specific messages based on JWT claims
- Admin-only broadcasts for sensitive data
- Rate limiting per client

## Performance

### Optimization

- Message batching for high-frequency updates
- Selective broadcasting (only to interested clients)
- Compression for large payloads
- Connection pooling

### Monitoring

- Track connected clients count
- Monitor message throughput
- Alert on connection failures
- Log reconnection attempts

## Testing

### Backend Tests

```go
func TestWebSocketHub(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    
    // Test client registration
    client := &Client{
        ID: "test-client",
        Hub: hub,
        Send: make(chan []byte, 256),
    }
    
    hub.register <- client
    time.Sleep(100 * time.Millisecond)
    
    assert.Equal(t, 1, hub.GetClientCount())
}
```

### Frontend Tests

```typescript
describe('useWebSocket', () => {
  it('should connect to WebSocket', () => {
    const { result } = renderHook(() => useWebSocket({
      url: 'ws://localhost:8080/ws',
    }));
    
    expect(result.current.isConnected).toBe(false);
    // Wait for connection
    waitFor(() => {
      expect(result.current.isConnected).toBe(true);
    });
  });
});
```

## Troubleshooting

### Connection Issues

1. **CORS errors**: Configure CORS in Gin server
2. **Authentication failures**: Check JWT token validity
3. **Connection drops**: Verify network stability
4. **Reconnection loops**: Check reconnection logic

### Performance Issues

1. **High latency**: Reduce message frequency
2. **Memory leaks**: Ensure proper cleanup
3. **CPU spikes**: Optimize message processing
4. **Network congestion**: Implement throttling

## Best Practices

1. **Graceful Degradation**: Fall back to polling if WebSocket fails
2. **Reconnection Strategy**: Exponential backoff with max attempts
3. **Message Validation**: Validate all incoming messages
4. **Error Handling**: Log errors, notify users
5. **Resource Cleanup**: Close connections properly
6. **Monitoring**: Track WebSocket metrics

## Future Enhancements

- [ ] Message compression (gzip)
- [ ] Binary message support
- [ ] Pub/sub channels
- [ ] Message persistence
- [ ] Load balancing across multiple servers
- [ ] WebSocket clustering

#