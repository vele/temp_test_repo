package event

import (
	"context"
	"sync"
)

type InMemoryPublisher struct {
	mu     sync.Mutex
	events []Event
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{}
}

func (p *InMemoryPublisher) Publish(_ context.Context, evt Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, evt)
	return nil
}

func (p *InMemoryPublisher) Events() []Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Event, len(p.events))
	copy(out, p.events)
	return out
}
