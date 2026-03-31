package sse

import (
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string][]chan string
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string][]chan string),
	}
}

func (h *Hub) Register(date string) chan string {
	ch := make(chan string, 4)
	h.mu.Lock()
	h.clients[date] = append(h.clients[date], ch)
	h.mu.Unlock()
	return ch
}

func (h *Hub) Deregister(date string, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.clients[date]
	for i, c := range list {
		if c == ch {
			h.clients[date] = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(h.clients[date]) == 0 {
		delete(h.clients, date)
	}
	close(ch)
}

func (h *Hub) Broadcast(date, payload string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.clients[date] {
		select {
		case ch <- payload:
		default:
		}
	}
}
