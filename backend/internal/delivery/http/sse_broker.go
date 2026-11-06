package http

import (
	"sync"
)

type SSEBroker struct {
	Clients map[chan string]bool
	mu      sync.Mutex
}

func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		Clients: make(map[chan string]bool),
	}
}

func (b *SSEBroker) AddClient(client chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Clients[client] = true
}

func (b *SSEBroker) RemoveClient(client chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Clients, client)
	close(client)
}

func (b *SSEBroker) Broadcast(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for client := range b.Clients {
		select {
		case client <- msg:
			// Sent successfully
		default:
			// Client channel is full/blocked, skip to prevent server freeze
		}
	}
}
