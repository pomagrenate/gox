package bus

import (
	"sync"
)

// LocalBus is a simple, memory-based implementation of EventBus using Go channels.
type LocalBus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(data any)
}

// NewLocalBus creates a new instance of LocalBus.
func NewLocalBus() *LocalBus {
	return &LocalBus{
		subscribers: make(map[string][]func(data any)),
	}
}

// Publish sends an event asynchronously to all registered subscribers.
func (b *LocalBus) Publish(topic string, data any) {
	b.mu.RLock()
	handlers, found := b.subscribers[topic]
	b.mu.RUnlock()

	if !found {
		return
	}

	for _, handler := range handlers {
		// Run each handler in a new goroutine to ensure asynchronous, non-blocking execution
		go func(h func(data any)) {
			h(data)
		}(handler)
	}
}

// Subscribe registers a new handler for a given topic.
func (b *LocalBus) Subscribe(topic string, handler func(data any)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.subscribers[topic] = append(b.subscribers[topic], handler)
}
