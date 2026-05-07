package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeInstanceUpdate MessageType = "instance_update"
	MessageTypeMetricsUpdate  MessageType = "metrics_update"
	MessageTypeAlert          MessageType = "alert"
	MessageTypeHeartbeat      MessageType = "heartbeat"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// Client represents a WebSocket client
type Client struct {
	ID       string
	Hub      *Hub
	Send     chan []byte
	UserID   int
	Username string
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	broadcast chan []byte

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket: Client %s registered (User: %s)", client.ID, client.Username)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("WebSocket: Client %s unregistered (User: %s)", client.ID, client.Username)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client's send channel is full, close it
					close(client.Send)
					delete(h.clients, client)
					log.Printf("WebSocket: Client %s send buffer full, disconnected", client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(messageType MessageType, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := Message{
		Type:      messageType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.broadcast <- msgBytes
	return nil
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID int, messageType MessageType, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := Message{
		Type:      messageType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- msgBytes:
			default:
				log.Printf("WebSocket: Failed to send to user %d, buffer full", userID)
			}
		}
	}

	return nil
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetClients returns a list of connected client IDs
func (h *Hub) GetClients() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]string, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client.ID)
	}
	return clients
}
