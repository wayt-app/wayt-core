package sse

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Hub manages SSE client connections.
// Each client is identified by a key (e.g. "customer:1", "owner:2").
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[int64]chan string
	seq     atomic.Int64
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[int64]chan string)}
}

// Subscribe registers a new channel for the given key and returns (channel, client_id).
func (h *Hub) Subscribe(key string) (chan string, int64) {
	id := h.seq.Add(1)
	ch := make(chan string, 8)
	h.mu.Lock()
	if h.clients[key] == nil {
		h.clients[key] = make(map[int64]chan string)
	}
	h.clients[key][id] = ch
	h.mu.Unlock()
	return ch, id
}

// Unsubscribe removes a client channel.
func (h *Hub) Unsubscribe(key string, id int64) {
	h.mu.Lock()
	if m, ok := h.clients[key]; ok {
		delete(m, id)
		if len(m) == 0 {
			delete(h.clients, key)
		}
	}
	h.mu.Unlock()
}

// Publish sends a message to all clients subscribed to key.
func (h *Hub) Publish(key, data string) {
	h.mu.RLock()
	m := h.clients[key]
	h.mu.RUnlock()
	for _, ch := range m {
		select {
		case ch <- data:
		default: // drop if buffer full
		}
	}
}

// CustomerKey returns the SSE key for a customer.
func CustomerKey(id uint) string { return fmt.Sprintf("customer:%d", id) }

// OwnerKey returns the SSE key for a business owner.
func OwnerKey(id uint) string { return fmt.Sprintf("owner:%d", id) }

// StaffKey returns the SSE key for a staff member.
func StaffKey(id uint) string { return fmt.Sprintf("staff:%d", id) }
