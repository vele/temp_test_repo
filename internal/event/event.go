package event

import (
	"context"
	"time"
)

type Type string

const (
	UserCreated Type = "UserCreated"
	UserUpdated Type = "UserUpdated"
	UserDeleted Type = "UserDeleted"
)

type Event struct {
	Type       Type        `json:"type"`
	UserID     uint        `json:"user_id"`
	Payload    interface{} `json:"payload"`
	OccurredAt time.Time   `json:"occurred_at"`
}

type Publisher interface {
	Publish(ctx context.Context, evt Event) error
}

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
}

type Handler func(ctx context.Context, evt Event) error
