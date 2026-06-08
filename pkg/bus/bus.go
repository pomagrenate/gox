package bus

// EventBus is the core interface for the internal event system.
// It enables loosely coupled communication between modules.
type EventBus interface {
	// Publish emits an event with the given topic and data payload.
	Publish(topic string, data any)
	// Subscribe registers a handler function for a specific topic.
	Subscribe(topic string, handler func(data any))
}
